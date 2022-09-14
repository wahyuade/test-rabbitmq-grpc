package schemas

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/rabbitmq/amqp091-go"
)

// in case we have more than one server id
const SERVER_ID = "main"

type Rpc interface {
	User() UserRPC
	Product() ProductRPC
	Transaction() TransactionRPC
	Order() OrderRPC
}

type rpc struct {
	user_rpc        user_rpc
	product_rpc     product_rpc
	transaction_rpc transaction_rpc
	order_rpc       order_rpc
	conn            *amqp091.Connection
}

func InitRpc() Rpc {
	conn, err := amqp091.Dial(os.Getenv("RABBITMQ_URI"))
	if err != nil {
		log.Panicf("Failed to connect to RabbitMQ: %s", err)
	}
	u := user_rpc{
		conn: conn,
	}
	p := product_rpc{
		conn: conn,
	}
	t := transaction_rpc{
		conn: conn,
	}
	o := order_rpc{
		conn: conn,
	}
	return rpc{
		user_rpc:        u,
		product_rpc:     p,
		transaction_rpc: t,
		order_rpc:       o,
	}
}

func (r rpc) User() UserRPC {
	return r.user_rpc
}

func (r rpc) Product() ProductRPC {
	return r.product_rpc
}

func (r rpc) Transaction() TransactionRPC {
	return r.transaction_rpc
}

func (r rpc) Order() OrderRPC {
	return r.order_rpc
}

func newChannel(conn *amqp091.Connection, topic string) *amqp091.Channel {
	ch, err := conn.Channel()
	if err != nil {
		log.Panicf("Failed to open a channel: %s", err)
	}
	ch.ExchangeDeclare(
		topic,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	ch.Qos(1, 0, true)
	return ch
}

func createListener(ch *amqp091.Channel, topic, callbackQueueName, uniqueRoutingKey string) <-chan amqp091.Delivery {
	q, err := ch.QueueDeclare(
		// fmt.Sprintf("%s_%s", SERVER_ID, callbackQueueName),
		uniqueRoutingKey,
		false,
		true,
		true,
		true,
		nil,
	)
	if err != nil {
		log.Panicf("Failed to declare queue: %s", err)
	}
	err = ch.QueueBind(
		q.Name,
		uniqueRoutingKey,
		topic,
		false,
		nil,
	)
	if err != nil {
		log.Panicf("Failed to bind queue: %s", err)
	}
	message, err := ch.Consume(
		q.Name,
		"callback",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Panicf("Failed to listen the callback: %s", err)
	}
	return message
}

func waitForResponse(ctx context.Context, ch *amqp091.Channel, topic string, queue string, callbackQueue string, correlationId string, payload []byte) []byte {
	callbackRoutingKey := fmt.Sprintf("%s-%s", callbackQueue, correlationId)
	response := createListener(ch, topic, callbackQueue, callbackRoutingKey)
	go ch.PublishWithContext(
		ctx,
		topic,
		queue,
		false,
		false,
		amqp091.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationId,
			ReplyTo:       callbackRoutingKey,
			Body:          payload,
		},
	)
	var delivery = <-response
	if delivery.CorrelationId == correlationId {
		delivery.Ack(false)
		return delivery.Body
	} else {
		delivery.Nack(false, true)
	}
	return nil
}
