package services

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"wahyuade.com/simple-e-commerce/schemas"
)

const TOPIC_PRODUCT = "product-service"
const CREATE_PRODUCT_QUEUE = "createProduct"
const LIST_PRODUCT_QUEUE = "listProduct"
const DETAIL_PRODUCT_QUEUE = "detailProduct"
const DECREMENT_PRODUCT_STOCK_QUEUE = "decrementProductStock"

type ProductService struct {
	conn   *amqp091.Connection
	dbConn *sql.DB
}

func (pS ProductService) Init(dbConn *sql.DB, conn *amqp091.Connection) ServiceRunner {
	return ProductService{
		conn:   conn,
		dbConn: dbConn,
	}
}

func (pS ProductService) Start() {
	pS.RegisterAllHandler()
}

func (pS ProductService) RegisterAllHandler() {
	go pS.createProductHandler()
	go pS.listProductHandler()
	go pS.detailProductHandler()
	go pS.decrementProductStockHandler()
}

func (pS ProductService) Name() string {
	return "product-service"
}

func (pS ProductService) createProductHandler() {
	log.Printf("%s Handler Registered\n", CREATE_PRODUCT_QUEUE)
	ch := newChannel(pS.conn, TOPIC_PRODUCT)
	msgs := listen(ch, TOPIC_PRODUCT, CREATE_PRODUCT_QUEUE)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", CREATE_PRODUCT_QUEUE, d.CorrelationId)
			product := schemas.Product{}
			json.Unmarshal(d.Body, &product)

			uuidV4 := uuid.NewString()
			product.Uuid = uuidV4

			row := pS.dbConn.QueryRow(
				`INSERT INTO "product"(uuid, name, description, stock, price) VALUES ($1, $2, $3, $4, $5) RETURNING 1`,
				product.Uuid,
				product.Name,
				product.Description,
				product.Stock,
				product.Price,
			)
			var int int
			err := row.Scan(&int)
			if err != nil {
				log.Println(err)
				product.Uuid = ""
			}
			resp, _ := json.Marshal(product)
			reply(ch, d, resp)
		}
	}()
	<-forever
}
func (pS ProductService) listProductHandler() {
	log.Printf("%s Handler Registered\n", LIST_PRODUCT_QUEUE)
	ch := newChannel(pS.conn, TOPIC_PRODUCT)
	msgs := listen(ch, TOPIC_PRODUCT, LIST_PRODUCT_QUEUE)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", LIST_PRODUCT_QUEUE, d.CorrelationId)
			var products []schemas.Product
			rows, err := pS.dbConn.Query(`SELECT uuid, name, price FROM "product"`)
			if err != nil {
				log.Println(err)
			}
			for rows.Next() {
				var product schemas.Product
				rows.Scan(&product.Uuid, &product.Name, &product.Price)
				products = append(products, product)
			}
			resp, _ := json.Marshal(products)
			reply(ch, d, resp)
		}
	}()
	<-forever
}
func (pS ProductService) detailProductHandler() {
	log.Printf("%s Handler Registered\n", DETAIL_PRODUCT_QUEUE)
	ch := newChannel(pS.conn, TOPIC_PRODUCT)
	msgs := listen(ch, TOPIC_PRODUCT, DETAIL_PRODUCT_QUEUE)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", DETAIL_PRODUCT_QUEUE, d.CorrelationId)
			productUuid := string(d.Body)
			product := schemas.Product{}
			row := pS.dbConn.QueryRow(
				`SELECT uuid, name, description, stock, price FROM "product" WHERE uuid = $1`,
				productUuid,
			)
			err := row.Scan(&product.Uuid, &product.Name, &product.Description, &product.Stock, &product.Price)
			if err != nil {
				log.Println(err)
			}
			resp, err := json.Marshal(product)
			if err != nil {
				log.Println(err)
			}
			reply(ch, d, resp)
		}
	}()
	<-forever
}

func (pS ProductService) decrementProductStockHandler() {
	log.Printf("%s Handler Registered\n", DECREMENT_PRODUCT_STOCK_QUEUE)
	ch := newChannel(pS.conn, TOPIC_PRODUCT)
	msgs := listen(ch, TOPIC_PRODUCT, DECREMENT_PRODUCT_STOCK_QUEUE)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", DECREMENT_PRODUCT_STOCK_QUEUE, d.CorrelationId)
			product := schemas.Product{}
			json.Unmarshal(d.Body, &product)

			tx, err := pS.dbConn.Begin()
			if err != nil {
				log.Println(err)
				reply(ch, d, nil)
				continue
			}
			row := tx.QueryRow(
				`UPDATE "product" SET stock = stock - $1 WHERE uuid = $2`,
				product.Stock,
				product.Uuid,
			)
			if row.Err() != nil {
				log.Println(err)
				tx.Rollback()
				reply(ch, d, []byte("false"))
				continue
			}
			tx.Commit()
			reply(ch, d, []byte("true"))
		}
	}()
	<-forever
}
