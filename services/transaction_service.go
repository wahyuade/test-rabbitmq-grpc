package services

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"wahyuade.com/simple-e-commerce/helpers"
	"wahyuade.com/simple-e-commerce/schemas"
)

const TOPIC_TRANSACTION = "transaction-service"

const CHECK_PRODUCT_IN_CART_QUEUE = "checkProductInCart"
const INSERT_PRODUCT_TO_CART_QUEUE = "insertProductToCart"
const LIST_CART_QUEUE = "listCart"
const LIST_TRANSACTION_ITEM_QUEUE = "listTransactionItem"
const UPDATE_CART_QUEUE = "updateCart"
const CREATE_BILLING_QUEUE = "createBilling"
const DETAIL_BILLING_QUEUE = "detailBilling"
const PROCESS_PAYMENT_QUEUE = "processPayment"

type TransactionService struct {
	conn   *amqp091.Connection
	dbConn *sql.DB
}

func (tS TransactionService) Init(dbConn *sql.DB, conn *amqp091.Connection) ServiceRunner {
	return TransactionService{
		conn:   conn,
		dbConn: dbConn,
	}
}

func (tS TransactionService) Start() {
	tS.RegisterAllHandler()
}

func (tS TransactionService) RegisterAllHandler() {
	go tS.checkProductInCartHandler()
	go tS.insertProductToCartHandler()
	go tS.listCartHandler()
	go tS.updateCartHandler()
	go tS.createBillingHandler()
	go tS.listTransactionItemHandler()
	go tS.detailBillingHandler()
	go tS.processPaymentHandler()
}

func (tS TransactionService) Name() string {
	return "transaction-service"
}

func (tS TransactionService) checkProductInCartHandler() {
	log.Printf("%s Handler Registered\n", CHECK_PRODUCT_IN_CART_QUEUE)
	ch := newChannel(tS.conn, TOPIC_TRANSACTION)
	msgs := listen(ch, TOPIC_TRANSACTION, CHECK_PRODUCT_IN_CART_QUEUE)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", CHECK_PRODUCT_IN_CART_QUEUE, d.CorrelationId)
			cart := schemas.TransactionCart{}
			json.Unmarshal(d.Body, &cart)

			tx, err := tS.dbConn.Begin()
			if err != nil {
				log.Println(err)
				reply(ch, d, nil)
				continue
			}
			row := tx.QueryRow(
				`SELECT user_uuid, product_uuid, qty FROM "transaction_cart" WHERE user_uuid = $1 AND product_uuid=$2 AND transaction_uuid IS NULL FOR UPDATE`,
				cart.UserUuid,
				cart.ProductUuid,
			)
			err = row.Scan(&cart.UserUuid, &cart.ProductUuid, &cart.Qty)
			if err != nil {
				cart.ProductUuid = ""
				resp, _ := json.Marshal(cart)
				log.Println(err)
				tx.Rollback()
				reply(ch, d, resp)
				continue
			}
			tx.Commit()
			resp, _ := json.Marshal(cart)
			reply(ch, d, resp)
		}
	}()
	<-forever
}

func (tS TransactionService) insertProductToCartHandler() {
	log.Printf("%s Handler Registered\n", INSERT_PRODUCT_TO_CART_QUEUE)
	ch := newChannel(tS.conn, TOPIC_TRANSACTION)
	msgs := listen(ch, TOPIC_TRANSACTION, INSERT_PRODUCT_TO_CART_QUEUE)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", INSERT_PRODUCT_TO_CART_QUEUE, d.CorrelationId)
			cart := schemas.TransactionCart{}
			json.Unmarshal(d.Body, &cart)

			tx, err := tS.dbConn.Begin()
			if err != nil {
				log.Println(err)
				reply(ch, d, nil)
				continue
			}
			uuidV4 := uuid.NewString()
			row := tx.QueryRow(
				`INSERT INTO "transaction_cart" (uuid, user_uuid, product_uuid, qty) VALUES ($1, $2, $3, $4) RETURNING 1`,
				uuidV4,
				cart.UserUuid,
				cart.ProductUuid,
				cart.Qty,
			)
			var status int
			err = row.Scan(&status)
			if err != nil {
				resp, _ := json.Marshal(cart)
				log.Println(err)
				tx.Rollback()
				reply(ch, d, resp)
				continue
			}
			tx.Commit()
			cart.Uuid = uuidV4
			resp, _ := json.Marshal(cart)
			reply(ch, d, resp)
		}
	}()
	<-forever
}

func (tS TransactionService) listCartHandler() {
	log.Printf("%s Handler Registered\n", LIST_CART_QUEUE)
	ch := newChannel(tS.conn, TOPIC_TRANSACTION)
	msgs := listen(ch, TOPIC_TRANSACTION, LIST_CART_QUEUE)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", LIST_CART_QUEUE, d.CorrelationId)
			var carts []schemas.TransactionCart
			rows, err := tS.dbConn.Query(`SELECT uuid, user_uuid, product_uuid, qty FROM "transaction_cart" WHERE user_uuid = $1 AND transaction_uuid IS NULL`, string(d.Body))
			if err != nil {
				log.Println(err)
			} else {
				for rows.Next() {
					var cart schemas.TransactionCart
					rows.Scan(&cart.Uuid, &cart.UserUuid, &cart.ProductUuid, &cart.Qty)
					carts = append(carts, cart)
				}
			}
			resp, err := json.Marshal(carts)
			if err != nil {
				log.Println(err)
			}
			reply(ch, d, resp)
		}
	}()
	<-forever
}

func (tS TransactionService) listTransactionItemHandler() {
	log.Printf("%s Handler Registered\n", LIST_TRANSACTION_ITEM_QUEUE)
	ch := newChannel(tS.conn, TOPIC_TRANSACTION)
	msgs := listen(ch, TOPIC_TRANSACTION, LIST_TRANSACTION_ITEM_QUEUE)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", LIST_TRANSACTION_ITEM_QUEUE, d.CorrelationId)
			var carts []schemas.TransactionCart
			rows, err := tS.dbConn.Query(`SELECT uuid, user_uuid, product_uuid, qty FROM "transaction_cart" WHERE transaction_uuid = $1`, string(d.Body))
			if err != nil {
				log.Println(err)
			} else {
				for rows.Next() {
					var cart schemas.TransactionCart
					rows.Scan(&cart.Uuid, &cart.UserUuid, &cart.ProductUuid, &cart.Qty)
					carts = append(carts, cart)
				}
			}
			resp, err := json.Marshal(carts)
			if err != nil {
				log.Println(err)
			}
			reply(ch, d, resp)
		}
	}()
	<-forever
}

func (tS TransactionService) updateCartHandler() {
	log.Printf("%s Handler Registered\n", UPDATE_CART_QUEUE)
	ch := newChannel(tS.conn, TOPIC_TRANSACTION)
	msgs := listen(ch, TOPIC_TRANSACTION, UPDATE_CART_QUEUE)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", UPDATE_CART_QUEUE, d.CorrelationId)
			cart := schemas.TransactionCart{}
			json.Unmarshal(d.Body, &cart)

			tx, err := tS.dbConn.Begin()
			if err != nil {
				log.Println(err)
				reply(ch, d, nil)
				continue
			}
			row := tx.QueryRow(
				`UPDATE "transaction_cart" SET qty = $1 WHERE product_uuid = $2 AND user_uuid = $3 AND transaction_uuid IS NULL RETURNING uuid, user_uuid, product_uuid, qty`,
				cart.Qty,
				cart.ProductUuid,
				cart.UserUuid,
			)
			err = row.Scan(&cart.Uuid, &cart.UserUuid, &cart.ProductUuid, &cart.Qty)
			if err != nil {
				log.Println(err)
				tx.Rollback()
				reply(ch, d, nil)
				continue
			}
			tx.Commit()
			resp, _ := json.Marshal(cart)
			reply(ch, d, resp)
		}
	}()
	<-forever
}

func (tS TransactionService) createBillingHandler() {
	log.Printf("%s Handler Registered\n", CREATE_BILLING_QUEUE)
	ch := newChannel(tS.conn, TOPIC_TRANSACTION)
	msgs := listen(ch, TOPIC_TRANSACTION, CREATE_BILLING_QUEUE)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", CREATE_BILLING_QUEUE, d.CorrelationId)

			transaction := schemas.Transaction{}
			json.Unmarshal(d.Body, &transaction)

			tx, err := tS.dbConn.Begin()
			if err != nil {
				log.Println(err)
				reply(ch, d, nil)
				continue
			}
			uuidV4 := uuid.NewString()
			row := tx.QueryRow(
				`INSERT INTO "transaction" (uuid, user_uuid, status, amount, payment_method, created, expired, virtual_account) VALUES (
					$1,
					$2,
					$3,
					$4,
					$5,
					$6,
					$7,
					$8
				) RETURNING uuid, user_uuid, status, amount, payment_method, created, expired, virtual_account`,
				uuidV4,
				transaction.UserUuid,
				"UNPAID",
				transaction.Amount,
				transaction.PaymentMethod,
				time.Now(),
				time.Now().Add(24*time.Hour),
				helpers.RandomNumber(16),
			)
			err = row.Scan(
				&transaction.Uuid,
				&transaction.UserUuid,
				&transaction.Status,
				&transaction.Amount,
				&transaction.PaymentMethod,
				&transaction.Created,
				&transaction.Expired,
				&transaction.VirtualAccount,
			)
			if err != nil {
				log.Println(err)
				tx.Rollback()
				reply(ch, d, nil)
				continue
			}
			row = tx.QueryRow(
				`UPDATE "transaction_cart" SET transaction_uuid = $1 WHERE user_uuid = $2 AND transaction_uuid IS NULL RETURNING 1`,
				transaction.Uuid,
				transaction.UserUuid,
			)
			var isOne int
			err = row.Scan(&isOne)
			if err != nil {
				log.Println(err)
				tx.Rollback()
				reply(ch, d, nil)
				continue
			}
			tx.Commit()
			resp, _ := json.Marshal(transaction)
			reply(ch, d, resp)
		}
	}()
	<-forever
}

func (tS TransactionService) detailBillingHandler() {
	log.Printf("%s Handler Registered\n", DETAIL_BILLING_QUEUE)
	ch := newChannel(tS.conn, TOPIC_TRANSACTION)
	msgs := listen(ch, TOPIC_TRANSACTION, DETAIL_BILLING_QUEUE)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", DETAIL_BILLING_QUEUE, d.CorrelationId)
			transaction := schemas.Transaction{}

			json.Unmarshal(d.Body, &transaction)

			row := tS.dbConn.QueryRow(
				`SELECT 
					uuid, user_uuid, status, amount, payment_method, created, expired, virtual_account
				FROM "transaction" WHERE payment_method = $1 AND virtual_account = $2 AND user_uuid = $3`,
				transaction.PaymentMethod,
				transaction.VirtualAccount,
				transaction.UserUuid,
			)

			err := row.Scan(
				&transaction.Uuid,
				&transaction.UserUuid,
				&transaction.Status,
				&transaction.Amount,
				&transaction.PaymentMethod,
				&transaction.Created,
				&transaction.Expired,
				&transaction.VirtualAccount,
			)
			if err != nil {
				log.Println(err)
				reply(ch, d, nil)
				continue
			}
			resp, _ := json.Marshal(transaction)
			reply(ch, d, resp)
		}
	}()
	<-forever
}

func (tS TransactionService) processPaymentHandler() {
	log.Printf("%s Handler Registered\n", PROCESS_PAYMENT_QUEUE)
	ch := newChannel(tS.conn, TOPIC_TRANSACTION)
	msgs := listen(ch, TOPIC_TRANSACTION, PROCESS_PAYMENT_QUEUE)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", PROCESS_PAYMENT_QUEUE, d.CorrelationId)
			var pf schemas.PaymentFlag
			json.Unmarshal(d.Body, &pf)

			tx, err := tS.dbConn.Begin()
			if err != nil {
				log.Println(err)
				reply(ch, d, nil)
				continue
			}

			row := tx.QueryRow(
				`UPDATE "transaction" SET status = 'PAID' WHERE uuid = $1 RETURNING 1`,
				pf.Transaction.Uuid,
			)
			var isSuccessUpdateStatus int
			err = row.Scan(&isSuccessUpdateStatus)
			if err != nil {
				log.Println(err)
				tx.Rollback()
				reply(ch, d, nil)
				continue
			}
			uuidV4 := uuid.NewString()
			row = tx.QueryRow(
				`INSERT INTO "transaction_payment" (uuid, transaction_uuid, reference, amount, payment_datetime) VALUES (
					$1,
					$2,
					$3,
					$4,
					$5
				) RETURNING uuid, transaction_uuid, reference, amount, payment_datetime`,
				uuidV4,
				pf.Transaction.Uuid,
				pf.Payment.Reference,
				pf.Payment.Amount,
				pf.Payment.PaymentDatetime,
			)
			err = row.Scan(&pf.Payment.Uuid, &pf.Payment.TransactionUuid, &pf.Payment.Reference, &pf.Payment.Amount, &pf.Payment.PaymentDatetime)
			if err != nil {
				log.Println(err)
				tx.Rollback()
				reply(ch, d, nil)
				continue
			}
			tx.Commit()
			resp, _ := json.Marshal(pf)
			reply(ch, d, resp)
		}
	}()
	<-forever
}
