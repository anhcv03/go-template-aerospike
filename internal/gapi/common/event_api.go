package gcommon

import (
	"context"
	"strconv"

	"google.golang.org/grpc/metadata"
)

/* Get eventAPI from GRPC context */
func GetEventAPIFromContext(ctx context.Context) bool {
	md, _ := metadata.FromIncomingContext(ctx)
	eventAPIStr := md.Get("eventAPI")
	if len(eventAPIStr) != 0 {
		eventAPI, err := strconv.ParseBool(eventAPIStr[0])
		if err != nil {
			return false
		}

		return eventAPI
	}

	return false
}
