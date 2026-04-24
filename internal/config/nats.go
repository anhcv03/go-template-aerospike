package config

type NATSConfig struct {
	Server string `mapstructure:"server"`
	ListenToSubject []string `mapstructure:"listen_to_subject"`
}
