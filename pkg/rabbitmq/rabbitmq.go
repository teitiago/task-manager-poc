package rabbitmq

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/streadway/amqp"
	"github.com/teitiago/task-manager-poc/internal/config"
	"go.uber.org/zap"
)

type rabbitMQConnection struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

var once sync.Once
var instance rabbitMQConnection

type rabbitMQBroker struct {
	*rabbitMQConnection
	queue    string
	exchange string
}

// getRabbitDSN Builds the RMQ DSN connection
func getRabbitDSN() string {
	addr := config.GetEnv("RMQ_ADDR", "127.0.0.1")
	user := config.GetEnv("RMQ_USER", "guest")
	pass := config.GetEnv("RMQ_PWD", "guest")
	port := config.GetEnv("RMQ_PORT", "5672")
	vHost := config.GetEnv("RMQ_VHOST", "tasks")

	return fmt.Sprintf(
		"amqp://%v:%v@%v:%v/%v",
		user,
		pass,
		addr,
		port,
		vHost,
	)
}

func NewRabbitMQBroker(exchange string, queue string) *rabbitMQBroker {

	once.Do(func() {
		// reuse the rabbitmq connection
		serverURL := getRabbitDSN()
		connection, err := amqp.Dial(serverURL)
		if err != nil {
			panic(err)
		}
		channel, err := connection.Channel()
		if err != nil {
			panic(err)
		}
		instance = rabbitMQConnection{connection: connection, channel: channel}

	})

	err := instance.channel.ExchangeDeclare(
		exchange, // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		panic(err)
	}

	return &rabbitMQBroker{rabbitMQConnection: &instance, exchange: exchange, queue: queue}

}

// Publish Publishes a given message on a broker given a specific routing key
func (b *rabbitMQBroker) Publish(payload interface{}, routing string, wg *sync.WaitGroup) {

	defer wg.Done()

	// convert body to json
	body, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error("error converting to json", zap.String("routing", routing), zap.Error(err))
	}

	zap.L().Debug("sending message to consumer", zap.String("routing", routing), zap.Any("payload", body))
	err = b.channel.Publish(
		b.exchange, // exchange
		routing,    // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         []byte(body),
			DeliveryMode: amqp.Persistent,
		})

	if err != nil {
		zap.L().Error("error sending the message", zap.String("routing", routing), zap.Error(err))
	}
}

// Consume consumes the messages from the queue
func (b *rabbitMQBroker) Consume(routing string, handler func(payload []byte) bool) {

	// create the queue if it doesn't already exist
	_, err := b.channel.QueueDeclare(b.queue, true, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	// bind the queue to the routing key
	err = b.channel.QueueBind(b.queue, routing, b.exchange, false, nil)
	if err != nil {
		panic(err)
	}

	msgs, err := b.channel.Consume(
		b.queue, // queue
		"",      // consumer
		false,   // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range msgs {
			zap.L().Debug("received message", zap.Any("payload", msg.Body), zap.String("routing", routing))

			var err error
			// ack msg based on handler process
			if handler(msg.Body) {
				err = msg.Ack(false)
			} else {
				err = msg.Nack(false, true)
			}

			if err != nil {
				zap.L().Error("consumer error", zap.Error(err))
			}
		}
		zap.L().Error("RMQ closed")
	}()

}

// Close closes the rabbitmq connections and channels
func (b *rabbitMQBroker) Close() {

	err := b.channel.Close()
	if err != nil {
		zap.L().Error("error closing channel", zap.Error(err))
	}
	err = b.connection.Close()
	if err != nil {
		zap.L().Error("error closing connection", zap.Error(err))
	}

}
