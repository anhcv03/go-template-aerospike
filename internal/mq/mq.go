package mq

import (
	"github.com/rs/zerolog/log"
	"github.com/wagslane/go-rabbitmq"
)

func Connection(url string) *rabbitmq.Conn {
	conn, err := rabbitmq.NewConn(url, rabbitmq.WithConnectionOptionsLogging)
	if err != nil {
		log.Fatal().Err(err).Msg("Can not connect to rabbitmq")
	}
	log.Info().Msgf("Start the rabbitmq connection %s", url)
	return conn
}
