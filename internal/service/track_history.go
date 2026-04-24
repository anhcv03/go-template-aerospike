package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	as "github.com/aerospike/aerospike-client-go/v8"
	astypes "github.com/aerospike/aerospike-client-go/v8/types"
	"github.com/rs/zerolog/log"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/model/dbmodel"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/model/httpmodel"
	pb "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/pkg/pb"
)

func (s *MainService) CreateTrackHistory(track *pb.Track) error {
	log.Debug().Msgf("enqueue track history id %d", track.ID)
	trackHistoryObj := &dbmodel.TrackHistory{
		TrackId:        track.ID,
		TrackNumber:    track.TrackNumber,
		Altitude:       track.Altitude,
		Latitude:       track.Latitude,
		Longitude:      track.Longitude,
		Velocity:       track.Velocity,
		Heading:		track.Heading,
		Identification: int32(*track.Identification),
		CreatedAt:      time.Now().UnixMilli(),
	}

	s.historyCh <- trackHistoryObj

	return nil
}

func (s *MainService) ProcessBatchHistory(ctx context.Context) {

	// log.Debug().Msg("process batch track history")
	// log.Debug().Int("batch size", s.BatchSize).
	// 	Int("batch max age", int(s.BatchMaxAge)).

	// log.Debug().Msgf("object in queue: ")
	// log.Debug().Msgf("enqueue track history id 2222")

	batch := make([]*dbmodel.TrackHistory, 0, s.BatchSize)
	timer := time.NewTimer(s.BatchMaxAge)
	defer timer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		log.Debug().Msgf("object in queue: %d batch size %d max_age: %d", len(s.historyCh), s.BatchSize, s.BatchMaxAge )
		log.Debug().Msg("insert batch track history")
		err := s.InsertBatchTrackHistory(ctx, batch)
		if err != nil {
			log.Error().Err(err).Msg("insert batch failed")
			batch = batch[:0]
			return
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case ev, ok := <-s.historyCh:
			if !ok {
				flush()
				return
			}
			batch = append(batch, ev)
			if len(batch) >= s.BatchSize {
				flush()
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(s.BatchMaxAge)
			}
		case <-timer.C:
			flush()
			timer.Reset(s.BatchMaxAge)
		}
	}
}

func (s *MainService) InsertBatchTrackHistory(ctx context.Context, batch []*dbmodel.TrackHistory) error {
	writePolicy := as.NewWritePolicy(0, 0)
	writePolicy.SendKey = true

	for _, item := range batch {
		if item == nil {
			continue
		}

		key, err := as.NewKey(s.SvcConfig.DbConfig.Namespace, trackHistorySetName, historyRecordKey(item.TrackId, item.CreatedAt))
		if err != nil {
			return err
		}

		err = s.DbClient.PutBins(writePolicy, key,
			as.NewBin(trackIDBinName, int64(item.TrackId)),
			as.NewBin("track_num", int64(item.TrackNumber)),
			as.NewBin("altitude", item.Altitude),
			as.NewBin("latitude", item.Latitude),
			as.NewBin("longitude", item.Longitude),
			as.NewBin("velocity", item.Velocity),
			as.NewBin("heading", item.Heading),
			as.NewBin("ident", int64(item.Identification)),
			as.NewBin("created_at", item.CreatedAt),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MainService) DeleteTrackHistory(trackId int32) error {
	log.Debug().Msgf("delete track history id %d", trackId)

	recordset, err := s.queryTrackHistoryByTrackID(trackId, false)
	if err != nil {
		log.Error().Err(err).Msgf("query track history id %d failed", trackId)
		return err
	}
	defer recordset.Close()

	writePolicy := as.NewWritePolicy(0, 0)

	for res := range recordset.Results() {
		if res.Err != nil {
			log.Error().Err(res.Err).Msgf("iterate track history id %d failed", trackId)
			return res.Err
		}

		if res.Record == nil || res.Record.Key == nil {
			continue
		}

		if _, err := s.DbClient.Delete(writePolicy, res.Record.Key); err != nil && !err.Matches(astypes.KEY_NOT_FOUND_ERROR) {
			log.Error().Err(err).Msgf("delete track history id %d failed", trackId)
			return err
		}
	}

	return nil
}

func (s *MainService) FindByTrackId(ctx context.Context, trackId, numOfHistory int) ([]httpmodel.TrackHistory, error) {
	log.Debug().Msgf("get history of track id %d", trackId)
	recordset, err := s.queryTrackHistoryByTrackID(int32(trackId), true)
	if err != nil {
		return nil, err
	}
	defer recordset.Close()

	result := make([]httpmodel.TrackHistory, 0, numOfHistory)
	for res := range recordset.Results() {
		if res.Err != nil {
			return nil, res.Err
		}
		if res.Record == nil {
			continue
		}

		item, err := historyFromBins(res.Record.Bins)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt > result[j].CreatedAt
	})

	if numOfHistory > 0 && len(result) > numOfHistory {
		result = result[:numOfHistory]
	}

	return result, nil
}

func (s *MainService) queryTrackHistoryByTrackID(trackID int32, includeBins bool) (*as.Recordset, error) {
	queryPolicy := as.NewQueryPolicy()
	queryPolicy.ExpectedDuration = as.SHORT
	queryPolicy.IncludeBinData = includeBins
	stmt := as.NewStatement(s.SvcConfig.DbConfig.Namespace, trackHistorySetName)
	stmt.SetFilter(as.NewEqualFilter(trackIDBinName, int64(trackID)))

	return s.DbClient.Query(queryPolicy, stmt)
}

func historyRecordKey(trackID int32, createdAt int64) string {
	return fmt.Sprintf("%d:%d", trackID, createdAt)
}

func historyFromBins(bins as.BinMap) (httpmodel.TrackHistory, error) {
	return httpmodel.TrackHistory{
		TrackId:        int32(readInt64Bin(bins, trackIDBinName)),
		TrackNumber:    int32(readInt64Bin(bins, "track_num")),
		Altitude:       readFloat32Bin(bins, "altitude"),
		Latitude:       readFloat32Bin(bins, "latitude"),
		Longitude:      readFloat32Bin(bins, "longitude"),
		Velocity:       readFloat32Bin(bins, "velocity"),
		Identification: int32(readInt64Bin(bins, "ident")),
		CreatedAt:      readInt64Bin(bins, "created_at"),
	}, nil
}

func readInt64Bin(bins as.BinMap, key string) int64 {
	value, ok := bins[key]
	if !ok || value == nil {
		return 0
	}

	switch v := value.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}

func readFloat32Bin(bins as.BinMap, key string) float32 {
	value, ok := bins[key]
	if !ok || value == nil {
		return 0
	}

	switch v := value.(type) {
	case float32:
		return v
	case float64:
		return float32(v)
	case int:
		return float32(v)
	case int8:
		return float32(v)
	case int16:
		return float32(v)
	case int32:
		return float32(v)
	case int64:
		return float32(v)
	case uint8:
		return float32(v)
	case uint16:
		return float32(v)
	case uint32:
		return float32(v)
	case uint64:
		return float32(v)
	default:
		return 0
	}
}
