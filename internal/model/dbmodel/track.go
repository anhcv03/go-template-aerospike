package dbmodel

type Track struct {
	ID             uint64 `gorm:"primaryKey"`     // ID cua bang db
	TrackId        int32  `gorm:"not null;index"` // Track Id
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
	CreatedAt      int64 `gorm:"autoCreateTime"` // Thoi gian tao
	UpdatedAt 	   int64 `gorm:"autoCreateTime"` // Thoi gian cap nhat
}

type TrackHistory struct {
	ID             uint64 `gorm:"primaryKey"`     // ID cua bang db
	TrackId        int32  `gorm:"not null;index"` // Track Id
	TrackNumber    int32
	Altitude       float32
	Latitude       float32
	Longitude      float32
	Heading		   float32
	Velocity       float32
	Identification int32
	CreatedAt      int64 `gorm:"autoCreateTime"` // Thoi gian tao
}
