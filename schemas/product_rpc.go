package schemas

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

const TOPIC_PRODUCT = "product-service"

const CREATE_PRODUCT_QUEUE = "createProduct"
const CREATE_PRODUCT_RSP_QUEUE = "createProductRsp"
const LIST_PRODUCT_QUEUE = "listProduct"
const LIST_PRODUCT_RSP_QUEUE = "listProductRsp"
const DETAIL_PRODUCT_QUEUE = "detailProduct"
const DETAIL_PRODUCT_RSP_QUEUE = "detailProductRsp"
const DECREMENT_PRODUCT_STOCK_QUEUE = "decrementProductStock"
const DECREMENT_PRODUCT_STOCK_RSP_QUEUE = "decrementProductStockRsp"

type ProductRPC interface {
	CreateProduct(ctx context.Context, correlationId string, product Product) (bool, Product)
	ListProduct(ctx context.Context, correlationId string) []Product
	DetailProduct(ctx context.Context, correlationId, productUuid string) (bool, Product)
	DecrementStock(ctx context.Context, correlationId string, product Product) bool
}

type product_rpc struct {
	conn *amqp091.Connection
}

func (p product_rpc) CreateProduct(ctx context.Context, correlationId string, product Product) (bool, Product) {
	requestPayload, _ := json.Marshal(product)
	ch := newChannel(p.conn, TOPIC_PRODUCT)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_PRODUCT, CREATE_PRODUCT_QUEUE, CREATE_PRODUCT_RSP_QUEUE, correlationId, requestPayload)
	responseProduct := Product{}
	if response != nil {
		json.Unmarshal(response, &responseProduct)
	}
	return responseProduct.Uuid != "", responseProduct
}

func (p product_rpc) ListProduct(ctx context.Context, correlationId string) (products []Product) {
	ch := newChannel(p.conn, TOPIC_PRODUCT)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_PRODUCT, LIST_PRODUCT_QUEUE, LIST_PRODUCT_RSP_QUEUE, correlationId, nil)
	if response != nil {
		json.Unmarshal(response, &products)
	}
	return products
}

func (p product_rpc) DetailProduct(ctx context.Context, correlationId, productUuid string) (status bool, product Product) {
	ch := newChannel(p.conn, TOPIC_PRODUCT)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_PRODUCT, DETAIL_PRODUCT_QUEUE, DETAIL_PRODUCT_RSP_QUEUE, correlationId, []byte(productUuid))
	if response != nil {
		json.Unmarshal(response, &product)
	}
	return product.Uuid != "", product
}

func (p product_rpc) DecrementStock(ctx context.Context, correlationId string, product Product) bool {
	requestPayload, _ := json.Marshal(product)
	ch := newChannel(p.conn, TOPIC_PRODUCT)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_PRODUCT, DECREMENT_PRODUCT_STOCK_QUEUE, DECREMENT_PRODUCT_STOCK_RSP_QUEUE, correlationId, requestPayload)
	if response != nil {
		json.Unmarshal(response, &product)
	}
	return string(response) == "true"
}
