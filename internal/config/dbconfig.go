package config

type PostgresqlConfig struct {
	//DB config
	DBHost string `mapstructure:"host"`
	DBPort int    `mapstructure:"port"`
	DBName string `mapstructure:"name"`
	DBUser string `mapstructure:"user"`
	DBPass string `mapstructure:"pass"`
}
