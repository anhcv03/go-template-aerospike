package util

import (
	"context"

	"github.com/rs/zerolog/log"
)

/***************************************************************************************************************/

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

/***************************************************************************************************************/
