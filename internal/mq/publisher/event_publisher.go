package publisher

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"

	rabbitmq "github.com/wagslane/go-rabbitmq"
)

type EventPublisher struct {
	Publisher    *rabbitmq.Publisher
	ExchangeName string
}

func NewEventPublisher(rbConn *rabbitmq.Conn, exchangeName string) *EventPublisher {

	publisher, err := rabbitmq.NewPublisher(
		rbConn,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName(exchangeName),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
		rabbitmq.WithPublisherOptionsExchangeKind("topic"),
		rabbitmq.WithPublisherOptionsExchangeDurable,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Can not start publisher!")
	}
	log.Info().Msgf("Start the publisher with exchange name %s", exchangeName)

	publisher.NotifyPublish(func(c rabbitmq.Confirmation) {
		log.Debug().Msgf("message confirmed from server. tag: %v, ack: %v", c.DeliveryTag, c.Ack)
	})

	return &EventPublisher{
		Publisher:    publisher,
		ExchangeName: exchangeName,
	}
}

func (ep *EventPublisher) Publish(ctx context.Context, data []byte, key []string) {
	err := ep.Publisher.PublishWithContext(ctx, data, key, rabbitmq.WithPublishOptionsExchange(ep.ExchangeName))
	if err != nil {
		log.Error().Err(err).Msg("Failed to publish event message with keys " + strings.Join(key, ","))
	}
	dataSize := len(data)
	log.Info().Msgf("Success to publish event message with key %s, %d", strings.Join(key, ","), dataSize)
}
