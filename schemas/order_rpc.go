package schemas

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

const TOPIC_ORDER = "order-service"

const LIST_ORDER_QUEUE = "listOrderQueue"
const LIST_ORDER_RSP_QUEUE = "listOrderQueueRsp"
const DETAIL_ORDER_QUEUE = "detailOrderQueue"
const DETAIL_ORDER_RSP_QUEUE = "detailOrderQueueRsp"
const PROCESS_ORDER_QUEUE = "processOrderQueue"
const PROCESS_ORDER_RSP_QUEUE = "processOrderQueueRsp"

type OrderRPC interface {
	ListOrder(ctx context.Context, correlationId string, userUuid string) []Order
	DetailOrder(ctx context.Context, correlationId string, order Order) Order
	ProcessOrder(ctx context.Context, correlationId string, order Order) Order
}

type order_rpc struct {
	conn *amqp091.Connection
}

func (o order_rpc) ListOrder(ctx context.Context, correlationId string, userUuid string) (orders []Order) {
	ch := newChannel(o.conn, TOPIC_ORDER)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_ORDER, LIST_ORDER_QUEUE, LIST_ORDER_RSP_QUEUE, correlationId, []byte(userUuid))
	if response != nil {
		json.Unmarshal(response, &orders)
		return orders
	}
	return orders
}
func (o order_rpc) DetailOrder(ctx context.Context, correlationId string, or Order) (order Order) {
	requestPayload, _ := json.Marshal(or)
	ch := newChannel(o.conn, TOPIC_ORDER)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_ORDER, DETAIL_ORDER_QUEUE, DETAIL_ORDER_RSP_QUEUE, correlationId, requestPayload)
	if response != nil {
		json.Unmarshal(response, &order)
		return order
	}
	return order
}
func (o order_rpc) ProcessOrder(ctx context.Context, correlationId string, or Order) (order Order) {
	requestPayload, _ := json.Marshal(or)
	ch := newChannel(o.conn, TOPIC_ORDER)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_ORDER, PROCESS_ORDER_QUEUE, PROCESS_ORDER_RSP_QUEUE, correlationId, requestPayload)
	if response != nil {
		json.Unmarshal(response, &order)
		return order
	}
	return order
}
