package service

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	config "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/config"
	gclient "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/gapi/client"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi/websocket"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/model/dbmodel"
	pb "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/pkg/pb"
	"gorm.io/gorm"

	moptions "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/go-co-op/gocron"
	"github.com/nats-io/nats.go"
	"github.com/nguyenngodinh/qmgo"
	"github.com/nguyenngodinh/qmgo/options"
	"github.com/rs/zerolog/log"
	socketio "github.com/vchitai/go-socket.io/v4"
)

const (
	ORDER_SERVICE = "order"
)

var db *qmgo.Database

var orderColl *qmgo.Collection

var rwm sync.RWMutex

func initColl() {
	orderColl = db.Collection(ORDER_SERVICE)

	createIndex(reflect.TypeOf(pb.Order{}), orderColl)
}

type MainService struct {
	DbClient       *gorm.DB
	gClient        *gclient.Client
	SvcConfig      *config.ServiceConfig
	scheduler      *gocron.Scheduler
	NATSConnection *nats.Conn
	SocketServer *socketio.Server
	httpClient     *http.Client
	OrderHub       *websocket.Hub
	trackCh   chan *dbmodel.Track
	trackMap		map[int32]*pb.Track
	historyCh   chan *dbmodel.TrackHistory
	BatchSize   int
	BatchMaxAge time.Duration
}

func createIndex(rType reflect.Type, collection *qmgo.Collection) {
	ctx := log.Logger.WithContext(context.Background())

	typeName := rType.Name()
	collectionName := collection.GetCollectionName()

	config.PrintDebugLog(ctx, "Create index for type: %s - Collection: %s", typeName, collectionName)

	for i := 0; i < rType.NumField(); i++ {
		tag := rType.Field(i).Tag

		compound := []string{
			tag.Get("bson"),
		}

		if tag.Get("compound_with") != "" {
			compound = append(compound, strings.Split(tag.Get("compound_with"), ",")...)
		}

		unique := tag.Get("index") == "unique"

		config.PrintDebugLog(ctx, "Compound: %+v - Unique: %v", compound, unique)

		if unique {
			collection.CreateOneIndex(context.Background(), options.IndexModel{
				Key:          compound,
				IndexOptions: &moptions.IndexOptions{Unique: &unique},
			})
		}
	}

	config.PrintDebugLog(ctx, "Done to create index for type: %s - Collection: %s", typeName, collectionName)
}

func (us *MainService) publishEvent(ctx context.Context, data []byte, routingKey string)  error {
	err := us.NATSConnection.Publish(routingKey, data)
	if err != nil {
		log.Debug().Msgf("Failed to publish to nats server for: %s", routingKey)
	} else {
		log.Debug().Msgf("Success to publish to nats server for: %s", routingKey)
	}
	return err
}

func New(dbClient *gorm.DB, cfg config.ServiceConfig, nc *nats.Conn, socketServer *socketio.Server) *MainService {
	// db = dbClient.Database(cfg.DbConfig.DBName)

	// initColl()

	svc := &MainService{
		DbClient:       dbClient,
		// gClient:        gc,
		SvcConfig:      &cfg,
		scheduler:      gocron.NewScheduler(time.UTC),
		NATSConnection: nc,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		SocketServer: socketServer,
		trackMap: make(map[int32]*pb.Track),
		trackCh: make (chan *dbmodel.Track, 10000),
		historyCh:   make(chan *dbmodel.TrackHistory, 10000),
		BatchSize:   cfg.TrackConfig.History_Batch_Size,
		BatchMaxAge: time.Duration(cfg.TrackConfig.History_Batch_Max_Age) * time.Millisecond,
	}

	// Cronjob: mỗi 1 phút tạo 1 đơn demo và dọn dẹp đơn demo cũ (>50)
	// Đã tắt cronjob tạo order tự động
	// svc.scheduler.Every(1).Minute().Do(func() {
	// 	ctx := log.Logger.WithContext(context.Background())
	// 	config.PrintDebugLog(ctx, "Demo cron: tick")
	// 	svc.RunDemoOrderCron(ctx)
	// })
	// svc.scheduler.StartAsync()

	return svc
}

// SetOrderHub sets the websocket hub for broadcasting order events
func (us *MainService) SetOrderHub(hub *websocket.Hub) {
	us.OrderHub = hub
}

func (s *MainService) MigrateDB() error {
	schemaName := "track_manager"
	createSchemaSQL := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)
	if err := s.DbClient.Exec(createSchemaSQL).Error; err != nil {
		log.Error().Err(err).Msg("create schema failed")
		return err
	}
	// ONLY USE FOR DEV + TEST, MUST MANUAL MIGRATE ON PRODUCTION USING EXTERNAL TOOL
	// s.DbClient.AutoMigrate(&dbmodel.MilitaryEquipment{}, &dbmodel.OtherModel)
	s.DbClient.AutoMigrate(
		&dbmodel.Track{},
		&dbmodel.TrackHistory{} )
	return nil
}