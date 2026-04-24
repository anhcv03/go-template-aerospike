package middleware

import (
	"context"
	"fmt"
	"time"

	config "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/config"
	common "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi/handlers/common"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggerMiddleware(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	var result any
	var err error

	defer func() {
		duration := time.Since(start).Milliseconds()
		method := info.FullMethod
		if method == "" {
			method = "-"
		}
		protocol := "GRPC"
		statusCode := fmt.Sprint(codes.Unknown)
		code, ok := status.FromError(err)
		if ok {
			statusCode = fmt.Sprint(code.Code())
		}

		ctx := log.Logger.WithContext(ctx)

		if err != nil {
			config.PrintErrorLog(
				ctx,
				err,
				"%s %s %s %d",
				method,
				protocol,
				statusCode,
				duration,
			)
		} else {
			config.PrintDebugLog(
				ctx,
				"%s %s %s %d",
				method,
				protocol,
				statusCode,
				duration,
			)
		}

		common.SetGRPCMetric(
			method,
			protocol,
			statusCode,
			duration,
		)
	}()

	result, err = handler(ctx, req)

	return result, err
}
