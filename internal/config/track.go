package config

type TrackConfig struct {
	History_Batch_Size    int `mapstructure:"history_batch_size"`
	History_Batch_Max_Age int `mapstructure:"history_batch_max_age"`
}
