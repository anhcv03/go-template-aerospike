package config

import (
	"context"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file

		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]

				break
			}
		}

		file = short

		return file + ":" + strconv.Itoa(line)
	}

	log.Logger = log.With().Caller().Logger()
}

type LoggerConfig struct {
	LogLevel           int    `mapstructure:"log_level"`
	RequestLevel       int    `mapstructure:"request_level"`
	PrettyPrintConsole bool   `mapstructure:"pretty_print_console"`
	MaxSizeMB          int    `mapstructure:"log_max_size_mb"`
	Compress           bool   `mapstructure:"log_compress"`
	LocalTime          bool   `mapstructure:"log_localtime"`
	MaxBackups         int    `mapstructure:"log_max_backups"`
	MaxAgeDays         int    `mapstructure:"log_max_age_day"`
	LogDir             string `mapstructure:"log_dir"`
	RotateDaily        bool   `mapstructure:"log_rotate_daily"`
}

/* Print error log */
func PrintErrorLog(
	ctx context.Context,
	err error,
	format string,
	v ...interface{},
) {
	log.Ctx(ctx).Error().Err(err).Msgf(format, v...)
}

/* Print fatal log */
func PrintFatalLog(
	ctx context.Context,
	err error,
	format string,
	v ...interface{},
) {
	log.Ctx(ctx).Fatal().Err(err).Msgf(format, v...)
}

/* Print warning log */
func PrintWarningLog(
	ctx context.Context,
	format string,
	v ...interface{},
) {
	log.Ctx(ctx).Warn().Msgf(format, v...)
}

/* Print debug log */
func PrintDebugLog(
	ctx context.Context,
	format string,
	v ...interface{},
) {
	log.Ctx(ctx).Debug().Msgf(format, v...)
}

/* Print info log */
func PrintInfoLog(
	ctx context.Context,
	format string,
	v ...interface{},
) {
	log.Ctx(ctx).Info().Msgf(format, v...)
}
