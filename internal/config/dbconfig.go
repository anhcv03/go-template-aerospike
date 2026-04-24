package config

type AerospikeConfig struct {
	Host                  string   `mapstructure:"host"`
	Port                  int      `mapstructure:"port"`
	Hosts                 []string `mapstructure:"hosts"`
	Namespace             string   `mapstructure:"namespace"`
	User                  string   `mapstructure:"user"`
	Pass                  string   `mapstructure:"pass"`
	ConnectionQueueSize   int      `mapstructure:"connection_queue_size"`
	MinConnectionsPerNode int      `mapstructure:"min_connections_per_node"`
	IdleTimeoutMs         int      `mapstructure:"idle_timeout_ms"`
	ConnectTimeoutMs      int      `mapstructure:"connect_timeout_ms"`
	TendIntervalMs        int      `mapstructure:"tend_interval_ms"`
	WarmUpConnections     int      `mapstructure:"warm_up_connections"`
	UseServicesAlternate  bool     `mapstructure:"use_services_alternate"`
}
