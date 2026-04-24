package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	config "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/config"
	gapi "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/gapi"
	hapi "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi"
	router "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi/router"
	service "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/service"
	postgresDB "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/db"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	socketio "github.com/vchitai/go-socket.io/v4"
	"github.com/vchitai/go-socket.io/v4/engineio"
	"github.com/vchitai/go-socket.io/v4/engineio/transport"
	"github.com/vchitai/go-socket.io/v4/engineio/transport/polling"
	"github.com/vchitai/go-socket.io/v4/engineio/transport/websocket"
)

const probeFlag string = "probe"

var serverCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the server",
	Long:  "Starts server",
	Run: func(cmd *cobra.Command, args []string) {
		runServer(args)
	},
}

func init() {
	serverCmd.Flags().BoolP(probeFlag, "p", false, "Probe readiness before startup.")

	rootCmd.AddCommand(serverCmd)
}

var allowOriginFunc = func(r *http.Request) bool {
	return true
}

func runServer(args []string) {
	ctx := log.Logger.WithContext(context.Background())

	/**
	* Load config file
	 */
	cfgFile := "."

	if len(args) != 0 {
		cfgFile = args[0]

		config.PrintDebugLog(ctx, "Use config file by argument: %+v", cfgFile)
	}

	config.PrintDebugLog(ctx, "Load config file: %s", cfgFile)

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		config.PrintFatalLog(ctx, err, "Failed to load config file: %s", cfgFile)

		os.Exit(1)
	}

	config.PrintDebugLog(ctx, "Config file content: %+v", cfg)

	/**
	* Setting logger
	 */
	if cfg.OtherConfig.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05 02-01-2006"})
	}

	// /**
	// * Setting tracer
	//  */
	// tp, err := initTracer(cfg.OtherConfig)
	// if err != nil {
	// 	config.PrintErrorLog(ctx, err, "Failed to init tracer")
	// }
	// defer func() {
	// 	if err := tp.Shutdown(context.Background()); err != nil {
	// 		config.PrintErrorLog(ctx, err, "Failed to shutdown tracer provider")
	// 	}
	// }()

	/**
	* Start mongoDB client connection
	 */
	// addr := fmt.Sprintf("mongodb://%s:%d/?replicaset=%s", cfg.DbConfig.DBHost, cfg.DbConfig.DBPort, cfg.DbConfig.DBReplica)
	// qmgoClient, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: addr})
	// if err != nil {
	// 	config.PrintFatalLog(ctx, err, "Failed to connect to MongoDB: %s", addr)

	// 	os.Exit(1)
	// } else {
	// 	config.PrintDebugLog(ctx, "Connected to connect to MongoDB: %s", addr)
	// }

	/**
	* Connect database
	 */
	pgd := postgresDB.NewPostgresConnection(cfg.DbConfig)
	db := pgd.ConnectDatabase()

	// /**
	// * Start RabbitMQ client connection
	//  */
	// uri := fmt.Sprintf("%s://%s:%s@%s:%d/%s",
	// 	cfg.RabbitmqConfig.Schema,
	// 	cfg.RabbitmqConfig.Username,
	// 	cfg.RabbitmqConfig.Password,
	// 	cfg.RabbitmqConfig.Host,
	// 	cfg.RabbitmqConfig.Port,
	// 	cfg.RabbitmqConfig.Vhost)
	// conn := mq.Connection(uri)
	// publisher := publisher.NewEventPublisher(conn, cfg.RabbitmqConfig.EventExchange)

	/**
	* Start NATS client connection
	 */
	natsClient, err := nats.Connect(cfg.NATSConfig.Server)
	if err != nil {
		config.PrintFatalLog(ctx, err, "Failed to connect NATs server: %s", cfg.NATSConfig.Server)
	} else {
		config.PrintDebugLog(ctx, "Connected to connect NATs server: %s", cfg.NATSConfig.Server)
	}
	defer natsClient.Drain()

	/**
	* Start GRPC client connection
	 */
	// grpcClient := gclient.New(cfg.GrpcConfig.GrpcChannels)

	// svc := service.New(qmgoClient, cfg, grpcClient, natsClient)


	// Start Socket Server

	type connectionContext struct {
		conn   socketio.Conn
		ctx    context.Context
		cancel context.CancelFunc
	}

	var (
		connMutex sync.RWMutex
		conns     = make(map[string]connectionContext)
	)

	socketServer := socketio.NewServer(&engineio.Options{
		// PingTimeout:  20 * time.Second,
		// PingInterval: 25 * time.Second,
		Transports: []transport.Transport{
			&polling.Transport{
				CheckOrigin: allowOriginFunc,
			},
			&websocket.Transport{
				CheckOrigin: allowOriginFunc,
			},
		},
	})

	cleanup := func(s socketio.Conn, reason string) {
		connMutex.Lock()
		connCtx, exists := conns[s.ID()]
		if exists {
			// Cancel all goroutines associated with this connection
			connCtx.cancel()
			delete(conns, s.ID())
		}
		connMutex.Unlock()

		log.Printf("Cleanup: %s, reason: %s", s.ID(), reason)

		// Leave all rooms
		s.LeaveAll()
	}

	socketServer.OnConnect("/", func(c socketio.Conn, m map[string]interface{}) error {
		log.Printf("Connected: %s", c.ID())
		log.Debug().Msgf("socket connection connected from local addr %s remote adrr %s", c.LocalAddr(), c.RemoteAddr())

		ctx, cancel := context.WithCancel(context.Background())

		connMutex.Lock()
		conns[c.ID()] = connectionContext{
			conn:   c,
			ctx:    ctx,
			cancel: cancel,
		}
		connMutex.Unlock()

		c.SetContext(ctx)
		c.Emit("connect ack!")

		return nil
	})
	socketServer.OnError("/", func(c socketio.Conn, err error) {
		// log.Err(err).Msgf("socket connection error from local addr %s remote adrr %s", c.LocalAddr(), c.RemoteAddr())
		// c.LeaveAll()
		// c.Close()
		cleanup(c, err.Error())
	})
	socketServer.OnDisconnect("/", func(c socketio.Conn, s string, m map[string]interface{}) {
		// log.Debug().Msgf("socket connection disconnected from local addr %s remote adrr %s", c.LocalAddr(), c.RemoteAddr())
		// c.LeaveAll()
		// c.Close()
		cleanup(c, err.Error())
	})

	go func() {
		err1 := socketServer.Serve()
		if err1 != nil {
			// log.Fatal().Msg("Failed to start socket server")
			os.Exit(1)
		}
	}()
	defer socketServer.Close()


	svc := service.New(db, cfg, natsClient, socketServer)
	errs := make(chan error, 2)

	/**
	* Start HTTP server
	 */
	config.PrintDebugLog(ctx, "Starting HTTP server...")
	httpServer := hapi.NewServer(svc, cfg)
	httpServer.InitI18n()
	router.Init(httpServer)
	go httpServer.Start(errs)

	/**
	* Start GRPC server
	 */
	config.PrintDebugLog(ctx, "Starting GRPC server...")

	grpcServer := gapi.NewServer(cfg, svc)
	go grpcServer.Start(errs)


	// ONLY USE FOR DEV + TEST, MUST MANUAL MIGRATE ON PRODUCTION USING EXTERNAL TOOL
	err = svc.MigrateDB()
	if err != nil {
		log.Fatal().Msg("Migrate DB failed")
	}

	/**
	Start process save Track History to DB
	**/
	go svc.ProcessBatchHistory(context.Background())

	/**
	*Start NATS subscriptions
	**/

	for i := range cfg.NATSConfig.ListenToSubject  {
		subj := cfg.NATSConfig.ListenToSubject[i]
		log.Debug().Msgf("NATS subscriptions subject %v", subj)
		go natsClient.QueueSubscribe(subj, config.GetModuleName(), func(m *nats.Msg) {
			log.Debug().Msgf("Receive msg %v from subj %v", m.Subject, subj)
			go svc.ReceiveNATSMsg(m)
		})
	}
	// Start broadcast track info via socket
	svc.AddBroadcastTrackWorker()


	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)

		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs

	config.PrintFatalLog(ctx, err, "Services terminate")
}
