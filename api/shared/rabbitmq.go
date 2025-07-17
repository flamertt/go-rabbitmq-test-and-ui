package shared

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

// NewRabbitMQ creates a new RabbitMQ connection
func NewRabbitMQ(rabbitmqURL string) (*RabbitMQ, error) {
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	return &RabbitMQ{
		Connection: conn,
		Channel:    ch,
	}, nil
}

// Close closes the RabbitMQ connection
func (r *RabbitMQ) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Connection != nil {
		r.Connection.Close()
	}
}

// SetupExchangeAndQueues sets up the necessary exchanges and queues
func (r *RabbitMQ) SetupExchangeAndQueues() error {
	// Declare the fanout exchange for order events
	err := r.Channel.ExchangeDeclare(
		"order_events_exchange", // name
		"fanout",                // type
		true,                    // durable
		false,                   // auto-deleted
		false,                   // internal
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %v", err)
	}

	// Declare queues for each service
	queues := []string{
		"payment_queue",
		"stock_reservation_queue",
		"shipping_queue",
		"order_status_queue",
	}

	for _, queueName := range queues {
		_, err := r.Channel.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %v", queueName, err)
		}

		// Bind queue to exchange
		err = r.Channel.QueueBind(
			queueName,               // queue name
			"",                      // routing key (empty for fanout)
			"order_events_exchange", // exchange
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s: %v", queueName, err)
		}
	}

	log.Println("RabbitMQ exchanges and queues setup completed")
	return nil
}

// PublishEvent publishes an event to the order_events_exchange
func (r *RabbitMQ) PublishEvent(event OrderEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	err = r.Channel.Publish(
		"order_events_exchange", // exchange
		"",                      // routing key
		false,                   // mandatory
		false,                   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish event: %v", err)
	}

	log.Printf("Published event: %s for order: %s", event.EventType, event.OrderID)
	return nil
}

// ConsumeEvents consumes events from a specific queue
func (r *RabbitMQ) ConsumeEvents(queueName string, handler func(OrderEvent) error) error {
	msgs, err := r.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack (we'll manually ack)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %v", err)
	}

	go func() {
		for d := range msgs {
			var event OrderEvent
			err := json.Unmarshal(d.Body, &event)
			if err != nil {
				log.Printf("Error unmarshaling event: %v", err)
				d.Nack(false, false) // send to dead letter queue
				continue
			}

			err = handler(event)
			if err != nil {
				log.Printf("Error handling event: %v", err)
				d.Nack(false, true) // requeue
				continue
			}

			d.Ack(false) // acknowledge successful processing
		}
	}()

	log.Printf("Started consuming from queue: %s", queueName)
	return nil
} 