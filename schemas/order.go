package schemas

import (
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/graphql-go/graphql"
)

type Order struct {
	path   string
	schema graphql.Schema
	r      Rpc
	ctx    *fiber.Ctx

	Uuid     string      `json:"uuid,omitempty"`
	UserUuid string      `json:"user_uuid,omitempty"`
	Amount   int         `json:"amount,omitempty"`
	Created  time.Time   `json:"created,omitempty"`
	Items    []OrderItem `json:"items"`
}

type OrderItem struct {
	OrderUuid   string `json:"order_uuid,omitempty"`
	ProductUuid string `json:"product_uuid,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Price       int    `json:"price,omitempty"`
	Qty         int    `json:"qty,omitempty"`
}

func orderType(name string, field []string) *graphql.Object {
	allField := map[string]*graphql.Field{
		"uuid":      typeString,
		"user_uuid": typeString,
		"amount":    typeInt,
		"created":   typeDatetime,
	}
	allowedField := getAllowedField(allField, field)
	return graphql.NewObject(
		graphql.ObjectConfig{
			Name:   name,
			Fields: *allowedField,
		},
	)
}

func orderItemType(name string, field []string) *graphql.Object {
	allField := map[string]*graphql.Field{
		"order_uuid":   typeString,
		"product_uuid": typeString,
		"name":         typeString,
		"description":  typeString,
		"price":        typeInt,
		"qty":          typeInt,
	}
	allowedField := getAllowedField(allField, field)
	return graphql.NewObject(
		graphql.ObjectConfig{
			Name:   name,
			Fields: *allowedField,
		},
	)
}

func (o Order) orderQuery() *graphql.Field {
	oT := orderType("order", []string{})
	oT.AddFieldConfig("items", &graphql.Field{
		Type: graphql.NewList(orderItemType("items", []string{})),
	})
	return &graphql.Field{
		Type: oT,
		Args: graphql.FieldConfigArgument{
			"uuid": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			correlationId := o.ctx.Locals("uuid").(string)
			user := o.ctx.Locals("user").(User)
			order := Order{
				Uuid:     p.Args["uuid"].(string),
				UserUuid: user.Uuid,
			}
			order = o.r.Order().DetailOrder(p.Context, correlationId, order)
			if order.Uuid == "" {
				return nil, errors.New("order tidak ditemukan")
			}
			return order, nil
		},
	}
}

func (o Order) ordersQuery() *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewList(orderType("orders", []string{})),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			correlationId := o.ctx.Locals("uuid").(string)
			user := o.ctx.Locals("user").(User)
			orders := o.r.Order().ListOrder(p.Context, correlationId, user.Uuid)
			return orders, nil
		},
	}
}

func newOrder(r Rpc) Order {
	order := Order{r: r}
	return order
}

func (o Order) Context(c *fiber.Ctx) Schema {
	order := Order{r: o.r, ctx: c}
	query := graphql.NewObject(graphql.ObjectConfig{
		Name: "OrderQuery",
		Fields: graphql.Fields{
			"order":  order.orderQuery(),
			"orders": order.ordersQuery(),
		},
	})
	s, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: query,
	})
	if err != nil {
		log.Fatalf("Gagal membuat schema, error: %v", err.Error())
	}
	order.schema = s
	return order
}

func (o Order) Path() string {
	return o.path
}

func (o Order) Config() *graphql.Schema {
	return &o.schema
}

func (o Order) Scope() string {
	return "private"
}

func (o Order) RPC() Rpc {
	return o.r
}
