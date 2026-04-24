package service

import (
	"context"
	"time"

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
	err := s.DbClient.WithContext(ctx).Create(&batch).Error

	return err
}

func (s *MainService) DeleteTrackHistory(trackId int32) error {
	log.Debug().Msgf("delete track history id %d", trackId)

	err := s.DbClient.Where("track_id = ?", trackId).Delete(&dbmodel.TrackHistory{}).Error
	if err != nil {
		log.Error().Err(err).Msgf("delete track history id %d failed", trackId)
		return err
	}

	return nil
}

func (s *MainService) FindByTrackId(ctx context.Context, trackId, numOfHistory int) ([]httpmodel.TrackHistory, error) {
	log.Debug().Msgf("get history of track id %d", trackId)
	var result []httpmodel.TrackHistory

	err := s.DbClient.Model(&dbmodel.TrackHistory{}).
		Select("*").Where("track_id = ?", trackId).
		Order("created_at desc").Limit(numOfHistory).
		Scan(&result).Error

	if err != nil {
		return result, err
	}

	return result, err
}
