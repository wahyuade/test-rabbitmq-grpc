package schemas

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

const TOPIC_TRANSACTION = "transaction-service"

const CHECK_PRODUCT_IN_CART_QUEUE = "checkProductInCart"
const CHECK_PRODUCT_IN_CART_RSP_QUEUE = "checkProductInCartRsp"
const INSERT_PRODUCT_TO_CART_QUEUE = "insertProductToCart"
const INSERT_PRODUCT_TO_CART_RSP_QUEUE = "insertProductToCartRsp"
const LIST_CART_QUEUE = "listCart"
const LIST_CART_RSP_QUEUE = "listCartRsp"
const UPDATE_CART_QUEUE = "updateCart"
const UPDATE_CART_RSP_QUEUE = "updateCartRsp"
const CREATE_BILLING_QUEUE = "createBilling"
const CREATE_BILLING_RSP_QUEUE = "createBillingRsp"
const LIST_TRANSACTION_ITEM_QUEUE = "listTransactionItem"
const LIST_TRANSACTION_ITEM_RSP_QUEUE = "listTransactionItemRsp"
const DETAIL_BILLING_QUEUE = "detailBilling"
const DETAIL_BILLING_RSP_QUEUE = "detailBillingRsp"
const PROCESS_PAYMENT_QUEUE = "processPayment"
const PROCESS_PAYMENT_RSP_QUEUE = "processPaymentRsp"

type TransactionRPC interface {
	CheckProductInCart(ctx context.Context, correlationId string, cart TransactionCart) TransactionCart
	InsertProductToCart(ctx context.Context, correlationId string, cart TransactionCart) TransactionCart
	ListCart(ctx context.Context, correlationId, userUuid string) []TransactionCart
	UpdateCart(ctx context.Context, correlationId string, cart TransactionCart) TransactionCart
	CreateBilling(ctx context.Context, correlationId string, trx Transaction) Transaction
	ListTransactionItem(ctx context.Context, correlationId, transactionUuid string) []TransactionCart
	DetailBilling(ctx context.Context, correlationId string, trx Transaction) Transaction
	ProcessPayment(ctx context.Context, correlationId string, pF PaymentFlag) PaymentFlag
}

type transaction_rpc struct {
	conn *amqp091.Connection
}

func (t transaction_rpc) CheckProductInCart(ctx context.Context, correlationId string, cart TransactionCart) (txCart TransactionCart) {
	requestPayload, _ := json.Marshal(cart)
	ch := newChannel(t.conn, TOPIC_TRANSACTION)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_TRANSACTION, CHECK_PRODUCT_IN_CART_QUEUE, CHECK_PRODUCT_IN_CART_RSP_QUEUE, correlationId, requestPayload)
	if response != nil {
		json.Unmarshal(response, &txCart)
	}
	return txCart
}

func (t transaction_rpc) InsertProductToCart(ctx context.Context, correlationId string, cart TransactionCart) (txCart TransactionCart) {
	requestPayload, _ := json.Marshal(cart)
	ch := newChannel(t.conn, TOPIC_TRANSACTION)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_TRANSACTION, INSERT_PRODUCT_TO_CART_QUEUE, INSERT_PRODUCT_TO_CART_RSP_QUEUE, correlationId, requestPayload)
	if response != nil {
		json.Unmarshal(response, &txCart)
	}
	return txCart
}

func (t transaction_rpc) UpdateCart(ctx context.Context, correlationId string, cart TransactionCart) (txCart TransactionCart) {
	requestPayload, _ := json.Marshal(cart)
	ch := newChannel(t.conn, TOPIC_TRANSACTION)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_TRANSACTION, UPDATE_CART_QUEUE, UPDATE_CART_RSP_QUEUE, correlationId, requestPayload)
	if response != nil {
		json.Unmarshal(response, &txCart)
	}
	return txCart
}

func (t transaction_rpc) ListCart(ctx context.Context, correlationId, userUuid string) (carts []TransactionCart) {
	ch := newChannel(t.conn, TOPIC_TRANSACTION)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_TRANSACTION, LIST_CART_QUEUE, LIST_CART_RSP_QUEUE, correlationId, []byte(userUuid))
	if response != nil {
		json.Unmarshal(response, &carts)
	}
	return carts
}

func (t transaction_rpc) CreateBilling(ctx context.Context, correlationId string, trx Transaction) (tx Transaction) {
	requestPayload, _ := json.Marshal(trx)
	ch := newChannel(t.conn, TOPIC_TRANSACTION)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_TRANSACTION, CREATE_BILLING_QUEUE, CREATE_BILLING_RSP_QUEUE, correlationId, requestPayload)
	if response != nil {
		json.Unmarshal(response, &tx)
	}
	return tx
}

func (t transaction_rpc) ListTransactionItem(ctx context.Context, correlationId, transactionUuid string) (carts []TransactionCart) {
	ch := newChannel(t.conn, TOPIC_TRANSACTION)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_TRANSACTION, LIST_TRANSACTION_ITEM_QUEUE, LIST_TRANSACTION_ITEM_RSP_QUEUE, correlationId, []byte(transactionUuid))
	if response != nil {
		json.Unmarshal(response, &carts)
	}
	return carts
}

func (t transaction_rpc) DetailBilling(ctx context.Context, correlationId string, trx Transaction) (tx Transaction) {
	requestPayload, _ := json.Marshal(trx)
	ch := newChannel(t.conn, TOPIC_TRANSACTION)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_TRANSACTION, DETAIL_BILLING_QUEUE, DETAIL_BILLING_RSP_QUEUE, correlationId, requestPayload)
	if response != nil {
		json.Unmarshal(response, &tx)
	}
	return tx
}

func (t transaction_rpc) ProcessPayment(ctx context.Context, correlationId string, pF PaymentFlag) (pFResponse PaymentFlag) {
	requestPayload, _ := json.Marshal(pF)
	ch := newChannel(t.conn, TOPIC_TRANSACTION)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_TRANSACTION, PROCESS_PAYMENT_QUEUE, PROCESS_PAYMENT_RSP_QUEUE, correlationId, requestPayload)
	if response != nil {
		json.Unmarshal(response, &pFResponse)
	}
	return pFResponse
}
