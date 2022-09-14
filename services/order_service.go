package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"wahyuade.com/simple-e-commerce/schemas"
)

const TOPIC_ORDER = "order-service"

const LIST_ORDER_QUEUE = "listOrderQueue"
const DETAIL_ORDER_QUEUE = "detailOrderQueue"
const PROCESS_ORDER_QUEUE = "processOrderQueue"

type OrderService struct {
	conn   *amqp091.Connection
	dbConn *sql.DB
}

func (oS OrderService) Init(dbConn *sql.DB, conn *amqp091.Connection) ServiceRunner {
	return OrderService{
		conn:   conn,
		dbConn: dbConn,
	}
}

func (oS OrderService) Start() {
	oS.RegisterAllHandler()
}

func (oS OrderService) RegisterAllHandler() {
	go oS.listOrderHandler()
	go oS.detailOrderHandler()
	go oS.processOrderHandler()
}

func (oS OrderService) Name() string {
	return "order-service"
}

func (oS OrderService) listOrderHandler() {
	log.Printf("%s Handler Registered\n", LIST_ORDER_QUEUE)
	ch := newChannel(oS.conn, TOPIC_ORDER)
	msgs := listen(ch, TOPIC_ORDER, LIST_ORDER_QUEUE)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", LIST_ORDER_QUEUE, d.CorrelationId)
			var orders []schemas.Order
			rows, err := oS.dbConn.Query(
				`SELECT uuid, user_uuid, amount, created FROM "order" WHERE user_uuid = $1`,
				string(d.Body),
			)
			if err != nil {
				log.Println(err)
			} else {
				for rows.Next() {
					var order schemas.Order
					rows.Scan(&order.Uuid, &order.UserUuid, &order.Amount, &order.Created)
					orders = append(orders, order)
				}
			}
			resp, err := json.Marshal(orders)
			if err != nil {
				log.Println(err)
			}
			reply(ch, d, resp)
		}
	}()
	<-forever
}
func (oS OrderService) detailOrderHandler() {
	log.Printf("%s Handler Registered\n", DETAIL_ORDER_QUEUE)
	ch := newChannel(oS.conn, TOPIC_ORDER)
	msgs := listen(ch, TOPIC_ORDER, DETAIL_ORDER_QUEUE)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", DETAIL_ORDER_QUEUE, d.CorrelationId)
			order := schemas.Order{}
			json.Unmarshal(d.Body, &order)

			row := oS.dbConn.QueryRow(
				`SELECT uuid, user_uuid, amount, created FROM "order" WHERE user_uuid = $1 AND uuid = $2`,
				order.UserUuid,
				order.Uuid,
			)

			err := row.Scan(&order.Uuid, &order.UserUuid, &order.Amount, &order.Created)
			if err != nil {
				log.Println(err)
				reply(ch, d, nil)
				continue
			}
			rows, err := oS.dbConn.Query(
				`SELECT order_uuid, product_uuid, name, description, price, qty FROM "order_item" WHERE order_uuid = $1`,
				order.Uuid,
			)
			if err != nil {
				log.Println(err)
				reply(ch, d, nil)
				continue
			}
			for rows.Next() {
				var ord schemas.OrderItem
				rows.Scan(&ord.OrderUuid, &ord.ProductUuid, &ord.Name, &ord.Description, &ord.Price, &ord.Qty)
				order.Items = append(order.Items, ord)
			}
			resp, _ := json.Marshal(order)
			reply(ch, d, resp)
		}
	}()
	<-forever
}
func (oS OrderService) processOrderHandler() {
	log.Printf("%s Handler Registered\n", PROCESS_ORDER_QUEUE)
	ch := newChannel(oS.conn, TOPIC_ORDER)
	msgs := listen(ch, TOPIC_ORDER, PROCESS_ORDER_QUEUE)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", PROCESS_ORDER_QUEUE, d.CorrelationId)

			order := schemas.Order{}
			json.Unmarshal(d.Body, &order)

			tx, err := oS.dbConn.Begin()
			if err != nil {
				log.Println(err)
				reply(ch, d, nil)
				continue
			}
			uuidV4 := uuid.NewString()
			row := tx.QueryRow(
				`INSERT INTO "order" (uuid, user_uuid, amount, created) VALUES (
					$1,
					$2,
					$3,
					NOW()
				) RETURNING uuid, user_uuid, amount, created`,
				uuidV4,
				order.UserUuid,
				order.Amount,
			)
			err = row.Scan(&order.Uuid, &order.UserUuid, &order.Amount, &order.Created)
			if err != nil {
				log.Println(err)
				tx.Rollback()
				reply(ch, d, nil)
				continue
			}
			var values []interface{}
			var binding []string

			values = append(values, uuidV4)
			for _, o := range order.Items {
				var tmpBinding []string
				tmpBinding = append(tmpBinding, "$1")
				values = append(values, o.ProductUuid)
				tmpBinding = append(tmpBinding, fmt.Sprintf("$%d", len(values)))
				values = append(values, o.Name)
				tmpBinding = append(tmpBinding, fmt.Sprintf("$%d", len(values)))
				values = append(values, o.Description)
				tmpBinding = append(tmpBinding, fmt.Sprintf("$%d", len(values)))
				values = append(values, o.Price)
				tmpBinding = append(tmpBinding, fmt.Sprintf("$%d", len(values)))
				values = append(values, o.Qty)
				tmpBinding = append(tmpBinding, fmt.Sprintf("$%d", len(values)))
				bundled := fmt.Sprintf("(%s)", strings.Join(tmpBinding, ","))
				binding = append(binding, bundled)
			}
			// bulk insert
			row = tx.QueryRow(
				fmt.Sprintf(`INSERT INTO "order_item" (order_uuid, product_uuid, name, description, price, qty) VALUES %s`, strings.Join(binding, ",")),
				values...,
			)
			if row.Err() != nil {
				log.Println(row.Err())
				tx.Rollback()
				reply(ch, d, nil)
				continue
			}
			tx.Commit()
			resp, _ := json.Marshal(order)
			reply(ch, d, resp)
		}
	}()
	<-forever
}
