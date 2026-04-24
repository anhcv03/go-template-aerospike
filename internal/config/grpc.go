package config

type GrpcConfig struct {
	GrpcPort     int               `mapstructure:"port"`
	GrpcHost     string            `mapstructure:"host"`
	GrpcChannels map[string]string `mapstructure:"channels"`
}
