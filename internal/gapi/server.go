package gapi

import (
	"context"
	"fmt"
	"net"

	config "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/config"
	logger "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/gapi/middleware"
	service "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/service"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	Config      config.ServiceConfig
	MainService *service.MainService
}

func NewServer(cfg config.ServiceConfig, svc *service.MainService) *Server {
	return &Server{
		Config:      cfg,
		MainService: svc,
	}
}

func (s *Server) Start(errs chan error) {
	ctx := log.Logger.WithContext(context.Background())

	grpcLogger := grpc.UnaryInterceptor(logger.LoggerMiddleware)

	// embedded logger to grpc server
	grpcServer := grpc.NewServer(grpcLogger)

	// orderHandler := order.NewOrderServiceHandler(s.MainService)
	// pb.RegisterOrderServiceServer(grpcServer, orderHandler)

	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", fmt.Sprintf("%v:%d", s.Config.GrpcConfig.GrpcHost, s.Config.GrpcConfig.GrpcPort))
	if err != nil {
		config.PrintFatalLog(ctx, err, "Cannot create grpc listener")

		errs <- err
	}

	config.PrintDebugLog(ctx, "Start GRPC server on: %s", listener.Addr().String())

	err = grpcServer.Serve(listener)
	if err != nil {
		config.PrintFatalLog(ctx, err, "Cannot start grpc server")

		errs <- err
	}
}
