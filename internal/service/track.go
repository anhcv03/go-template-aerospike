package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/model/dbmodel"
	pb "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/pkg/pb"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/pkg/pb/ipc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)
var (
	Fixed_Wing_Shape   = "Cánh Bằng"
	Rotary_Wing_Shape  = "Cánh Quạt"
	Unknown_Wing_Shape = "CXD"
)

const (
	UnknownShape = iota
	FixedWingShape
	RotaryWingShape
)

var workers = cron.New()

func (s *MainService) AddBroadcastTrackWorker() {
	log.Debug().Msg("Add func to worker")
	workers.AddFunc("@every 1s", s.broadcastTrackInfo)
}

func (s *MainService) ProcessIpcEvent(msgType ipc.MessageType, msgAny *anypb.Any) {
	log.Debug().Msgf("Process ipc event msg type %v", msgType)
	switch msgType {
	// BAN TIN CAP NHAT TUNG QUY DAO TU ipc
	case ipc.MessageType_MT_TRACK_MESSAGE:
		s.processIpcTrackUpdate(msgAny)
	// BAN TIN XOA 1 LIST QUY DAO TU ipc
	case ipc.MessageType_MT_REMOVED_TRACKS_EVENT_MESSAGE:
		s.processIpcTrackRemove(msgAny)
	// BAN TIN STATISTIC CHUA ID CUA CAC TRACK DANG TON TAI
	case ipc.MessageType_MT_TRACK_STATISTICS_MESSAGE:
		s.processIpcTrackStatistic(msgAny)
	default:
		log.Debug().Msgf("Ignore process ipc event msg type %v", msgType)
	}
}

func (s *MainService) processIpcTrackUpdate(msgAny *anypb.Any) {
	log.Debug().Msg("process ipc track update")
	tMsg := ipc.TrackMessage{}
	msgAny.UnmarshalTo(&tMsg)
	track, err := convertIpcTrackToTrack(&tMsg)
	if err != nil {
		log.Error().Err(err)
	}

	data, err := proto.Marshal(track)
	if err != nil {
		log.Error().Err(err).Msg("marshal proto failed")
		return
	}

	routingKey := "track.update"
	err = s.publishEvent(context.Background(), data, routingKey)
	if err != nil {
		log.Error().Err(err).Msg("publish track update message failed")
	}
	// Update Track Into Map
	s.updateTrackToMap(track)

	// CREATE track history
	s.CreateTrackHistory(track)

	// Update track info to db
	// s.createTrackInfo(track);
}

func (s *MainService) updateTrackToMap(track *pb.Track) error {
	rwm.Lock()
	defer rwm.Unlock()
	s.trackMap[track.ID] = track;
	return nil
}

func (s *MainService) removeTrackFromMap(trackId int32) error {
	rwm.Lock()
	defer rwm.Unlock()
	delete (s.trackMap, trackId)
	return nil
}

func (s *MainService) removeTracksFromMap(trackIds []int32) error {
	rwm.Lock()
	defer rwm.Unlock()
	for idx := range trackIds {
		delete (s.trackMap, trackIds[idx])
	}
	return nil
}

func (s *MainService) createTrackInfo(track *pb.Track ) error {
	log.Debug().Msgf("enqueue track id %d", track.ID)
	trackObj := &dbmodel.Track {
		TrackId:        track.ID, // Track Id
		TrackNumber:    track.TrackNumber,
		Altitude:       track.Altitude,
		Latitude:       track.Latitude,
		Longitude:      track.Longitude,
		AircraftType:   int32(*track.AircraftType),
		Identification: int32(*track.Identification),
		Quantity:       track.Quantity,
		Model:          track.Model,
		Velocity:       track.Velocity,
		Heading:        track.Heading,
		Quality:        track.Quality,
		Mode3a:         track.Mode3A,
	}
	s.trackCh <- trackObj
	return nil
}

func (s *MainService) processIpcTrackRemove(msgAny *anypb.Any) {
	rMsg := ipc.RemovedTracksEventMessage{}
	msgAny.UnmarshalTo(&rMsg)
	log.Debug().Msgf("process ipc track remove %v", rMsg.String())
	for _, trackId := range rMsg.RemovedTrackIdList {
		deleteTrackMsg := pb.DeleteTrack{
			TrackId: trackId,
		}

		data, err := proto.Marshal(&deleteTrackMsg)
		if err != nil {
			log.Error().Err(err).Msg("marshal proto failed")
			return
		}

		routingKey := "track.delete"
		err = s.publishEvent(context.Background(), data, routingKey)
		if err != nil {
			log.Error().Err(err).Msg("publish track delete message failed")
		}

		// Delete track from track map
		s.removeTrackFromMap(trackId)

		//DELETE track history
		log.Debug().Msgf("delete track history trackId: %v", trackId)
		s.DeleteTrackHistory(trackId)

		// s.trackAlert.Delete(trackId)
	}
}

func (s *MainService) processIpcTrackStatistic(msgAny *anypb.Any) {
	sMsg := ipc.TrackStatisticsMessage{}
	msgAny.UnmarshalTo(&sMsg)
	log.Debug().Msgf("process ipc track statistic %v", sMsg.String())
	listTrackMsg := pb.ListCurrentTracks{
		TrackId: sMsg.AllIdList,
	}

	data, err := proto.Marshal(&listTrackMsg)
	if err != nil {
		log.Error().Err(err).Msg("marshal proto failed")
		return
	}

	trackIds := sMsg.GetAllIdList()
	mapTrackIds := s.GetAllTrackIds()
	deleteTrackIds := diffArray(mapTrackIds, trackIds)

	log.Debug().Msgf("process unique track statistic %v", sMsg.String())

	if(len(deleteTrackIds) > 0){
		s.removeTracksFromMap(deleteTrackIds)
	}

	routingKey := "track.all"
	err = s.publishEvent(context.Background(), data, routingKey)
	if err != nil {
		log.Error().Err(err).Msg("publish track statistic message failed")
	}
}

func diffArray[T comparable](a, b []T) []T {
	if len(a) == 0{
		return []T{}
	}

	setB := make(map[T]struct{}, len(b))
	result := make([]T, 0)
	for idx := range b {
		setB[b[idx]] = struct{}{}
	}
	
	for idx := range a {
		if _, ok := setB[a[idx]]; !ok {
			result = append(result, a[idx])
		}
	}

	return result;
}

func convertIpcTrackToTrack(ipcTrack *ipc.TrackMessage) (*pb.Track, error) {
	track := &pb.Track{}
	log.Debug().Msgf("convert ipc track: %v", ipcTrack)
	if ipcTrack != nil {
		if ipcTrack.Id != nil {
			track.ID = *ipcTrack.Id
			//temporary set track number is same as track id
			track.TrackNumber = *ipcTrack.Id
		} else {
			return track, errors.New("ipc track id is nil")
		}

		if ipcTrack.CreatedTimestamp != nil {
			track.CreatedAt = *ipcTrack.CreatedTimestamp
		}

		if ipcTrack.LastUpdatedTimestamp != nil {
			track.UpdatedAt = *ipcTrack.LastUpdatedTimestamp
		}

		if ipcTrack.SourceTracks != nil {
			for i := range ipcTrack.SourceTracks {
				st := ipcTrack.SourceTracks[i]
				tn := strconv.Itoa(int(*st.Id))
				track.ListSources = append(track.ListSources, &pb.TrackSourceInfo{
					SourceName:  *st.SourceInfo.Id,
					TrackNumber: tn,
				})
			}
		}

		if ipcTrack.TrackInfo != nil {
			if ipcTrack.TrackInfo.AircraftInfo != nil {
				aircraftType := pb.AIRCRAFT_TYPE_AT_UNKNOWN
				if ipcTrack.TrackInfo.AircraftInfo.AircraftType != nil {
					aircraftType = pb.AIRCRAFT_TYPE(*ipcTrack.TrackInfo.AircraftInfo.AircraftType)
				}
				track.AircraftType = &aircraftType
				if ipcTrack.TrackInfo.AircraftInfo.Mode_3A != nil {
					track.Mode3A = *ipcTrack.TrackInfo.AircraftInfo.Mode_3A
				}
				if ipcTrack.TrackInfo.AircraftInfo.AircraftCount != nil {
					track.Quantity = *ipcTrack.TrackInfo.AircraftInfo.AircraftCount
				}
				if ipcTrack.TrackInfo.AircraftInfo.Model != nil {
					track.Model = *ipcTrack.TrackInfo.AircraftInfo.Model
				}
				if ipcTrack.TrackInfo.AircraftInfo.ModeSerial != nil {
					track.Serial = *ipcTrack.TrackInfo.AircraftInfo.ModeSerial
				}
				if ipcTrack.TrackInfo.AircraftInfo.Type1 != nil {
					shapeInfo := ipcTrack.TrackInfo.AircraftInfo.Type1
					track.Shape = getShapeStr(int(*shapeInfo.Type), int(*shapeInfo.Proba*100))
				}
			} else {
				return track, errors.New("ipc aircraft info is nil")
			}
			if ipcTrack.TrackInfo.TrackQuality != nil {
				track.Quality = *ipcTrack.TrackInfo.TrackQuality
			}

			identification := pb.TRACK_IDENTIFICATION_TI_UNKNOWN
			if ipcTrack.TrackInfo.FriendFoeState != nil {
				identification = pb.TRACK_IDENTIFICATION(*ipcTrack.TrackInfo.FriendFoeState)
			}
			track.Identification = &identification
		} else {
			return track, errors.New("ipc track info is nil")
		}

		if ipcTrack.GeodeticPosition != nil {
			track.Altitude = *ipcTrack.GeodeticPosition.Altitude
			track.Longitude = *ipcTrack.GeodeticPosition.Longitude * (180 / math.Pi)
			track.Latitude = *ipcTrack.GeodeticPosition.Latitude * (180 / math.Pi)
		} else {
			return track, errors.New("ipc track geodetis info is nil")
		}

		if ipcTrack.PolarVelocity != nil {
			if ipcTrack.PolarVelocity.Speed != nil {
				track.Velocity = *ipcTrack.PolarVelocity.Speed
			}
			if ipcTrack.PolarVelocity.Heading != nil {
				track.Heading = *ipcTrack.PolarVelocity.Heading * (180 / math.Pi)
			}
		} else {
			return track, errors.New("ipc track polar velocity info is nil")
		}
	} else {
		return track, errors.New("ipc track is nil")
	}
	log.Debug().Msgf("converted track: %v %d", track, track.Identification)
	return track, nil
}

func getShapeStr(shapeType int, probability int) string {
	if probability >= 70 {
		switch shapeType {
		case FixedWingShape:
			return fmt.Sprintf("%s_%d", Fixed_Wing_Shape, probability)
		case RotaryWingShape:
			return fmt.Sprintf("%s_%d", Rotary_Wing_Shape, probability)
		default:
			return Unknown_Wing_Shape
		}
	}
	return Unknown_Wing_Shape
}

func (s *MainService) ProcessBatchTrack( ctx context.Context) {

}

func (s *MainService) FindTrackAll(ctx context.Context) ([]*pb.Track, error) {
	rs := []*pb.Track{}
	rwm.RLock()

	for _, v := range s.trackMap {
		if (v != nil) {
			rs = append(rs, v)
		}
	}
	rwm.RUnlock()
	return rs, nil
}


func (s *MainService) FindTrackByID(ctx context.Context, id int32) (*pb.Track, error) {
	rwm.RLock()
	rs := s.trackMap[id]
	rwm.RUnlock()
	return rs, nil
}

func (s *MainService) GetAllTrackIds() []int32{
	rwm.RLock()
	mapTrackIds := make([]int32, 0, len(s.trackMap))
	for i:= range s.trackMap {
		mapTrackIds = append(mapTrackIds, i)
	}
	rwm.RUnlock()
	return mapTrackIds
}

func (s *MainService) broadcastTrackInfo() {
	rs := []*pb.Track{}
	rwm.RLock()

	for _, v := range s.trackMap {
		if (v != nil) {
			rs = append(rs, v)
		}
	}
	rwm.RUnlock()
	log.Debug().Msgf("broadcast tracks via socket: %v tracks", len(rs))
	data, _ := json.Marshal(rs)
	s.SocketServer.BroadcastToNamespace("/", "tracks", string(data))
}