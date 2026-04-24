package gclient

import (
	"context"
	"fmt"

	config "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/config"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	ClientConnMap  map[string]*grpc.ClientConn
	serviceConnMap map[string]string
}

func New(targetMap map[string]string) *Client {
	ctx := log.Logger.WithContext(context.Background())

	ccm := map[string]*grpc.ClientConn{}

	for svc, uri := range targetMap {
		gc, err := grpc.NewClient(
			uri,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to connect to GRPC service: %s - Address: %s", svc, uri)
		} else {
			ccm[svc] = gc

			config.PrintDebugLog(ctx, "Success to connect to GRPC service: %s - Address: %s", svc, uri)
		}
	}

	return &Client{
		ClientConnMap:  ccm,
		serviceConnMap: targetMap,
	}
}

func (c *Client) GetConn(svc string) (*grpc.ClientConn, error) {
	_, ok := c.serviceConnMap[svc]
	if ok {
		gConn, gOk := c.ClientConnMap[svc]
		if gOk {
			return gConn, nil
		}
	}

	return nil, fmt.Errorf("invalid svc: %s", c.serviceConnMap[svc])
}
