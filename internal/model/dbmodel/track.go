package dbmodel

type Track struct {
	TrackId        int32
	TrackNumber    int32
	Altitude       float32
	Latitude       float32
	Longitude      float32
	AircraftType   int32
	Identification int32
	Quantity       int32
	Model          string
	Velocity       float32
	Heading        float32
	Quality        int32
	Mode3a         string
	CreatedAt      int64
	UpdatedAt      int64
}

type TrackHistory struct {
	TrackId        int32
	TrackNumber    int32
	Altitude       float32
	Latitude       float32
	Longitude      float32
	Heading        float32
	Velocity       float32
	Identification int32
	CreatedAt      int64
}
