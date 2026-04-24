package config

type MongoConfig struct {
	DBHost    string `mapstructure:"host"`
	DBPort    int    `mapstructure:"port"`
	DBName    string `mapstructure:"name"`
	DBUser    string `mapstructure:"user"`
	DBPass    string `mapstructure:"pass"`
	DBReplica string `mapstructure:"replica"`
}
