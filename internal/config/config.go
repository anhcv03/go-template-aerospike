package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

const TRACK_SUBJECT = "ipc-track"

type ServiceConfig struct {
	DbConfig     PostgresqlConfig `mapstructure:"postgres"`
	GrpcConfig   GrpcConfig       `mapstructure:"grpc"`
	HttpConfig   HttpConfig       `mapstructure:"http"`
	LoggerConfig LoggerConfig     `mapstructure:"logger"`
	OtherConfig  OtherConfig      `mapstructure:"other"`
	NATSConfig   NATSConfig       `mapstructure:"nats"`
	TrackConfig       TrackConfig             `mapstructure:"tracks"`
}

func LoadConfig(path string) (cfg ServiceConfig, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.AutomaticEnv()

	setDefaultValue()

	err = viper.ReadInConfig()
	if err != nil {
		log.Fatal().Msg("Load config fail")
		return
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal().Msg("Error when mapping config")
	}

	err = InitLog(cfg.LoggerConfig)
	if err != nil {
		// log.Fatal().Msg("Error when Init logx")
		panic(err)
	}
	return
}

func setDefaultValue() {
	viper.SetDefault("other.environment", "development")

	viper.SetDefault("postgres.host", "127.0.0.1")
	viper.SetDefault("postgres.port", 5432)
	viper.SetDefault("postgres.name", "vqc-pk")
	viper.SetDefault("postgres.user", "postgres")
	viper.SetDefault("postgres.pass", "123456")

	viper.SetDefault("http.port", 8080)
	viper.SetDefault("http.enable_recover_middleware", true)
	viper.SetDefault("http.enable_cors_middleware", true)
	viper.SetDefault("http.coverage_service_url", "http://127.0.0.1:30038")

	viper.SetDefault("grpc.port", 8081)

	viper.SetDefault("other.default_lang", "en")
	viper.SetDefault("other.bundle_dir_abs", "./web/i18n")

	viper.SetDefault("nats.listen_to_subject", []string{TRACK_SUBJECT})
	viper.SetDefault("data.init_data", false)
	viper.SetDefault("data.uav_image_base_dir", "./image/uav")
	viper.SetDefault("data.file_base_dir", "./data/file")

	viper.SetDefault("mtx.host", "http://127.0.0.1")
	viper.SetDefault("mtx.api_port", 9997)
	viper.SetDefault("mtx.rtsp_port", 8554)
	viper.SetDefault("mtx.playback_port", 9996)
	viper.SetDefault("mtx.record_enable", true)
	viper.SetDefault("mtx.record_path", "./recordings/")
	viper.SetDefault("mtx.record_part_duration", "10m")
	viper.SetDefault("mtx.record_max_part_size", "50M")
	viper.SetDefault("mtx.record_segment_duration", "1h")
	viper.SetDefault("mtx.record_delete_after", "2h")

}

func InitLog(logCfg LoggerConfig) error {
	fmt.Print("InitLog:\n", logCfg.LogDir)
	if err := os.MkdirAll(resolveLogDir(logCfg.LogDir), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", logCfg.LogDir, err)
	}

	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logCfg.LogDir, "app.log"),
		MaxSize:    logCfg.MaxSizeMB,
		MaxBackups: logCfg.MaxBackups,
		MaxAge:     logCfg.MaxAgeDays,
		Compress:   logCfg.Compress,
		LocalTime:  logCfg.LocalTime,
	}

	if logCfg.RotateDaily {
		go rotateAtMidnight(fileWriter, time.Local)
	}

	var consoleWriter zerolog.ConsoleWriter
	if logCfg.PrettyPrintConsole {
		consoleWriter = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	} else {
		consoleWriter = zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: time.RFC3339}
		consoleWriter.FormatMessage = func(i interface{}) string {
			return fmt.Sprint(i)
		}
	}

	mw := zerolog.MultiLevelWriter(consoleWriter, fileWriter)

	zerolog.TimeFieldFormat = time.RFC3339Nano // Can change to other format
	zerolog.DurationFieldUnit = time.Millisecond

	log.Logger = zerolog.New(mw).
		With().Timestamp().
		Str("service", "utm-track-manager").Logger()

	return nil
}

func rotateAtMidnight(lj *lumberjack.Logger, loc *time.Location) {
	for {
		now := time.Now().In(loc)
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
		time.Sleep(time.Until(next))
		_ = lj.Rotate()
	}
}

func resolveLogDir(explicit string) string {
	if explicit != "" {
		return explicit
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", currentUser(), "logs")
}

func currentUser() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		return u.Username
	}
	if n := os.Getenv("USER"); n != "" {
		return n
	}

	return "user"
}

