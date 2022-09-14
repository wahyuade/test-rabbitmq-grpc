package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/rabbitmq/amqp091-go"
	"wahyuade.com/simple-e-commerce/database"
)

const SERVER_ID = "main"

type Service interface {
	Bootstrap(runner ServiceRunner)
}

type ServiceRunner interface {
	Name() string
	Init(*sql.DB, *amqp091.Connection) ServiceRunner
	Start()
	RegisterAllHandler()
}

type service struct {
	/**
		TODO: Maybe in the future we want to run multiple instance of the ServiceRunner
	**/
	conn   *amqp091.Connection
	dbConn *sql.DB
	worker int
}

func (s service) Bootstrap(runner ServiceRunner) {
	log.Printf("Bootstraping the %s...\n", runner.Name())
	svc := runner.Init(s.dbConn, s.conn)
	svc.Start()
}

func New() Service {
	uri := os.Getenv("RABBITMQ_URI")
	if uri == "" {
		uri = "amqp://guest:guest@localhost:5672/"
	}
	conn, err := amqp091.Dial(uri)
	if err != nil {
		log.Panicf("Failed to connect to RabbitMQ: %s", err)
	}
	log.Println("Successfully connected to RabbitMQ Server !")
	dbConnection := database.InitDB()
	return service{
		conn:   conn,
		dbConn: dbConnection,
	}
}

func listen(ch *amqp091.Channel, topic, queue string) <-chan amqp091.Delivery {
	q, _ := ch.QueueDeclare(
		fmt.Sprintf("%s_%s", SERVER_ID, queue),
		true,
		false,
		false,
		false,
		nil,
	)

	ch.QueueBind(
		q.Name,
		queue,
		topic,
		false,
		nil,
	)

	msgs, _ := ch.Consume(
		q.Name,
		fmt.Sprintf("%sWorker", queue),
		false,
		false,
		false,
		false,
		nil,
	)
	return msgs
}

func reply(ch *amqp091.Channel, d amqp091.Delivery, payload []byte) {
	ch.PublishWithContext(
		context.Background(),
		d.Exchange,
		d.ReplyTo,
		false,
		false,
		amqp091.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Body:          payload,
		},
	)
	d.Ack(false)
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
	return ch
}
