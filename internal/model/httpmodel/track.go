package httpmodel

import "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/pkg/pb"

type TrackHistory struct {
	TrackId        int32   `json:"track_id"`
	TrackNumber    int32   `json:"track_number"`
	Altitude       float32 `json:"altitude"`
	Latitude       float32 `json:"latitude"`
	Longitude      float32 `json:"longitude"`
	Velocity       float32 `json:"velocity"`
	Identification int32   `json:"identification"`
	CreatedAt      int64   `json:"created_at"`
}

type TracksListData struct {
	Items  []*pb.Track 				  `json:"items"`
	Limit  int                        `json:"limit" example:"20"`
	Offset int                        `json:"offset" example:"0"`
	Total  int64                      `json:"total" example:"4"`
}

type TrackHistoriesListData struct {
	Items  []TrackHistory 			  `json:"items"`
	Limit  int                        `json:"limit" example:"20"`
	Offset int                        `json:"offset" example:"0"`
	Total  int64                      `json:"total" example:"4"`
}

type TracksListResponse struct {
	Code    string                            `json:"code" example:"0"`
	Data    TracksListData 					  `json:"data"`
	Message string                            `json:"message" example:""`
}

type TrackResponse struct {
	Code    string                            `json:"code" example:"0"`
	Data    *pb.Track  					      `json:"data"`
	Message string                            `json:"message" example:""`
}

type TrackHistoriesListResponse struct {
	Code    string                            `json:"code" example:"0"`
	Data    TrackHistoriesListData 			  `json:"data"`
	Message string                            `json:"message" example:""`
}