package schemas

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/graphql-go/graphql"
)

type Transaction struct {
	path   string
	schema graphql.Schema
	r      Rpc
	ctx    *fiber.Ctx

	Uuid           string            `json:"uuid,omitempty"`
	UserUuid       string            `json:"user_uuid,omitempty"`
	Status         string            `json:"status,omitempty"`
	Amount         int               `json:"amount,omitempty"`
	PaymentMethod  string            `json:"payment_method,omitempty"`
	Created        time.Time         `json:"created,omitempty"`
	Expired        time.Time         `json:"expired,omitempty"`
	VirtualAccount string            `json:"virtual_account,omitempty"`
	Items          []TransactionCart `json:"items,omitempty"`
}

type TransactionCart struct {
	Uuid            string  `json:"uuid,omitempty"`
	UserUuid        string  `json:"user_uuid,omitempty"`
	ProductUuid     string  `json:"product_uuid,omitempty"`
	TransactionUuid string  `json:"transaction_uuid,omitempty"`
	Qty             int     `json:"qty,omitempty"`
	Product         Product `json:"product,omitempty"`
}

type TransactionPayment struct {
	Uuid            string      `json:"uuid,omitempty"`
	TransactionUuid string      `json:"transaction_uuid,omitempty"`
	Reference       string      `json:"reference,omitempty"`
	Amount          int         `json:"amount,omitempty"`
	PaymentDatetime time.Time   `json:"payment_datetime,omitempty"`
	Transaction     Transaction `json:"transaction,omitempty"`
}

type PaymentFlag struct {
	Transaction Transaction        `json:"transaction,omitempty"`
	Payment     TransactionPayment `json:"payment,omitempty"`
}

var typeStatusEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "Status",
	Values: graphql.EnumValueConfigMap{
		"UNPAID": &graphql.EnumValueConfig{
			Value: "UNPAID",
		},
		"PAID": &graphql.EnumValueConfig{
			Value: "PAID",
		},
	},
})
var typePaymentMethod = graphql.NewEnum(graphql.EnumConfig{
	Name: "PaymentMethod",
	Values: graphql.EnumValueConfigMap{
		"BCA": &graphql.EnumValueConfig{
			Value: "BCA",
		},
		"BNI": &graphql.EnumValueConfig{
			Value: "BNI",
		},
		"BRI": &graphql.EnumValueConfig{
			Value: "BRI",
		},
	},
})

func transactionType(name string, field []string) *graphql.Object {
	allField := map[string]*graphql.Field{
		"uuid":      typeString,
		"user_uuid": typeString,
		"status": &graphql.Field{
			Type: typeStatusEnum,
		},
		"payment_method": &graphql.Field{
			Type: typePaymentMethod,
		},
		"amount":          typeInt,
		"virtual_account": typeString,
		"created":         typeDatetime,
		"expired":         typeDatetime,
	}
	allowedField := getAllowedField(allField, field)
	return graphql.NewObject(
		graphql.ObjectConfig{
			Name:   name,
			Fields: *allowedField,
		},
	)
}

func transactionCartType(name string, field []string) *graphql.Object {
	allField := map[string]*graphql.Field{
		"uuid":         typeString,
		"user_uuid":    typeString,
		"product_uuid": typeString,
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

func transactionPaymentType(name string, field []string) *graphql.Object {
	allField := map[string]*graphql.Field{
		"uuid":             typeString,
		"reference":        typeString,
		"amount":           typeInt,
		"transaction_uuid": typeString,
		"payment_datetime": typeDatetime,
	}
	allowedField := getAllowedField(allField, field)
	return graphql.NewObject(
		graphql.ObjectConfig{
			Name:   name,
			Fields: *allowedField,
		},
	)
}

func (t Transaction) cartsQuery() *graphql.Field {
	carts := transactionCartType("carts", []string{})
	carts.AddFieldConfig("product", &graphql.Field{
		Type: productType("cart_product", []string{}),
	})
	return &graphql.Field{
		Type: graphql.NewList(carts),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			user := t.ctx.Locals("user").(User)
			correlationId := t.ctx.Locals("uuid").(string)
			carts := t.r.Transaction().ListCart(p.Context, correlationId, user.Uuid)
			for i, cart := range carts {
				_, product := t.r.Product().DetailProduct(p.Context, correlationId, cart.ProductUuid)
				carts[i].Product = product
			}
			return carts, nil
		},
	}
}

func (t Transaction) billingQuery() *graphql.Field {
	billing := transactionType("billing", []string{})
	item := transactionCartType("item", []string{})
	item.AddFieldConfig("product", &graphql.Field{
		Type: productType("product", []string{}),
	})
	billing.AddFieldConfig("items", &graphql.Field{
		Type: graphql.NewList(item),
	})
	return &graphql.Field{
		Type: billing,
		Args: graphql.FieldConfigArgument{
			"payment_method": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(typePaymentMethod),
			},
			"virtual_account": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			correlationId := t.ctx.Locals("uuid").(string)
			user := t.ctx.Locals("user").(User)
			paymentMethod := p.Args["payment_method"].(string)
			virtualAccount := p.Args["virtual_account"].(string)

			trx := Transaction{
				UserUuid:       user.Uuid,
				PaymentMethod:  paymentMethod,
				VirtualAccount: virtualAccount,
			}
			trx = t.r.Transaction().DetailBilling(p.Context, correlationId, trx)
			trx.Items = t.r.Transaction().ListTransactionItem(p.Context, correlationId, trx.Uuid)
			for i, it := range trx.Items {
				_, trx.Items[i].Product = t.r.Product().DetailProduct(p.Context, correlationId, it.ProductUuid)
			}
			if trx.Uuid == "" {
				return nil, errors.New("tagihan tidak ditemukan")
			}
			return trx, nil
		},
	}
}

func (t Transaction) addToCartMutation() *graphql.Field {
	return &graphql.Field{
		Type:        transactionCartType("addToCart", []string{"user_uuid", "product_uuid", "qty"}),
		Description: "Tambahkan produk ke kerajang",
		Args: graphql.FieldConfigArgument{
			"product_uuid": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"qty": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			productUuid := p.Args["product_uuid"].(string)
			qty := p.Args["qty"].(int)
			correlationId := t.ctx.Locals("uuid").(string)
			user := t.ctx.Locals("user").(User)
			status, product := t.r.Product().DetailProduct(context.Background(), correlationId, productUuid)
			if !status {
				return nil, errors.New("produk tidak tersedia")
			}

			if qty > product.Stock {
				return nil, errors.New("stock produk melebihi qty yang anda pesan")
			}
			if qty <= 0 {
				return nil, errors.New("qty harus lebih dari 0")
			}
			cart := TransactionCart{
				ProductUuid: productUuid,
				UserUuid:    user.Uuid,
				Qty:         qty,
			}
			checkCart := t.r.Transaction().CheckProductInCart(p.Context, correlationId, cart)
			if checkCart.ProductUuid == cart.ProductUuid {
				return nil, errors.New("produk sudah ada dalam keranjang")
			}

			insertedCart := t.r.Transaction().InsertProductToCart(p.Context, correlationId, cart)
			insertedCart.Product = product
			return insertedCart, nil
		},
	}
}

func (t Transaction) updateCartMutation() *graphql.Field {
	return &graphql.Field{
		Type:        transactionCartType("updateCart", []string{}),
		Description: "Kelola keranjang",
		Args: graphql.FieldConfigArgument{
			"product_uuid": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"qty": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			productUuid := p.Args["product_uuid"].(string)
			qty := p.Args["qty"].(int)
			correlationId := t.ctx.Locals("uuid").(string)
			user := t.ctx.Locals("user").(User)
			status, product := t.r.Product().DetailProduct(context.Background(), correlationId, productUuid)
			if !status {
				return nil, errors.New("produk tidak tersedia")
			}

			if qty > product.Stock {
				return nil, errors.New("stock produk melebihi qty yang anda pesan")
			}
			if qty <= 0 {
				return nil, errors.New("qty harus lebih dari 0")
			}
			cart := TransactionCart{
				ProductUuid: productUuid,
				UserUuid:    user.Uuid,
				Qty:         qty,
			}
			checkCart := t.r.Transaction().CheckProductInCart(p.Context, correlationId, cart)
			if checkCart.ProductUuid == "" {
				return nil, errors.New("produk tidak ada dalam keranjang")
			}

			updatedCart := t.r.Transaction().UpdateCart(p.Context, correlationId, cart)
			if updatedCart.Uuid == "" {
				return nil, errors.New("internal server error")
			}
			return updatedCart, nil
		},
	}
}

func (t Transaction) createBillingMutation() *graphql.Field {
	return &graphql.Field{
		Type:        transactionType("createBilling", []string{}),
		Description: "Pembuatan tagihan",
		Args: graphql.FieldConfigArgument{
			"payment_method": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(typePaymentMethod),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			correlationId := t.ctx.Locals("uuid").(string)
			user := t.ctx.Locals("user").(User)
			carts := t.r.Transaction().ListCart(p.Context, correlationId, user.Uuid)
			if len(carts) == 0 {
				return nil, errors.New("keranjang masih kosong")
			}
			amount := 0
			for _, cart := range carts {
				_, product := t.r.Product().DetailProduct(p.Context, correlationId, cart.ProductUuid)
				amount = amount + product.Price
			}

			transaction := Transaction{
				UserUuid:      user.Uuid,
				Amount:        amount,
				PaymentMethod: p.Args["payment_method"].(string),
			}
			trx := t.r.Transaction().CreateBilling(p.Context, correlationId, transaction)
			if trx.Uuid == "" {
				return nil, errors.New("internal server error")
			}
			return trx, nil
		},
	}
}

func (t Transaction) paymentMutation() *graphql.Field {
	payment := transactionPaymentType("payment", []string{})
	payment.AddFieldConfig("transaction", &graphql.Field{
		Type: transactionType("transaction_payment", []string{}),
	})
	return &graphql.Field{
		Type:        payment,
		Description: "Proses flag transaction",
		Args: graphql.FieldConfigArgument{
			"payment_method": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(typePaymentMethod),
			},
			"virtual_account": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"reference": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"amount": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"payment_datetime": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			correlationId := t.ctx.Locals("uuid").(string)
			user := t.ctx.Locals("user").(User)
			paymentMethod := p.Args["payment_method"].(string)
			virtualAccount := p.Args["virtual_account"].(string)
			reference := p.Args["reference"].(string)
			amount := p.Args["amount"].(int)
			paymentDatetime := p.Args["payment_datetime"].(string)
			trx := Transaction{
				UserUuid:       user.Uuid,
				PaymentMethod:  paymentMethod,
				VirtualAccount: virtualAccount,
			}
			trx = t.r.Transaction().DetailBilling(p.Context, correlationId, trx)
			if trx.Uuid == "" {
				return nil, errors.New("tagihan tidak ditemukan")
			}
			if trx.Status == "PAID" {
				return nil, errors.New("tagihan sudah terbayar")
			}

			if amount < trx.Amount {
				return nil, errors.New("uang anda kurang untuk melakukan pembayaran")
			}
			paymentDatetimeAsTime, err := time.Parse("2006-01-02 15:04:05", paymentDatetime)
			if err != nil {
				log.Println(err)
				return nil, errors.New("format payment_datetime harus YYYY-MM-DD HH:mm:ss contoh 2022-01-20 10:00:00")
			}
			paymentFlag := PaymentFlag{
				Transaction: trx,
				Payment: TransactionPayment{
					Reference:       reference,
					Amount:          amount,
					PaymentDatetime: paymentDatetimeAsTime,
				},
			}

			// implementasi saga harusnya disini
			pFlagRes := t.r.Transaction().ProcessPayment(p.Context, correlationId, paymentFlag)
			// proses order
			order := Order{
				UserUuid: user.Uuid,
				Amount:   trx.Amount,
			}
			transactionItem := t.r.Transaction().ListTransactionItem(p.Context, correlationId, trx.Uuid)
			for _, i := range transactionItem {
				// decrement stock
				_, product := t.r.Product().DetailProduct(p.Context, correlationId, i.ProductUuid)
				product.Stock = i.Qty
				go t.r.Product().DecrementStock(p.Context, correlationId, product)
				or := OrderItem{
					ProductUuid: i.ProductUuid,
					Name:        product.Name,
					Description: product.Description,
					Price:       product.Price,
					Qty:         i.Qty,
				}
				order.Items = append(order.Items, or)
			}
			order = t.r.Order().ProcessOrder(p.Context, correlationId, order)
			if order.Uuid == "" {
				return nil, errors.New("internal server error")
			}

			if pFlagRes.Payment.Uuid == "" {
				return nil, errors.New("internal server error")
			}

			pFlagRes.Payment.Transaction = trx
			return pFlagRes.Payment, nil
		},
	}
}

func newTransaction(rpc Rpc) Transaction {
	transaction := Transaction{
		r: rpc,
	}
	return transaction
}

func (t Transaction) Context(c *fiber.Ctx) Schema {
	transaction := Transaction{
		r:   t.r,
		ctx: c,
	}
	query := graphql.NewObject(graphql.ObjectConfig{
		Name: "TransactionQuery",
		Fields: graphql.Fields{
			"carts":   transaction.cartsQuery(),
			"billing": transaction.billingQuery(),
		},
	})
	mutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "TransactionMutation",
		Fields: graphql.Fields{
			"addToCart":     transaction.addToCartMutation(),
			"updateCart":    transaction.updateCartMutation(),
			"createBilling": transaction.createBillingMutation(),
			"payment":       transaction.paymentMutation(),
		},
	})
	s, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    query,
		Mutation: mutation,
	})
	if err != nil {
		log.Fatalf("Gagal membuat schema, error: %v", err.Error())
	}
	transaction.schema = s
	return transaction
}

func (t Transaction) Path() string {
	return t.path
}

func (t Transaction) Config() *graphql.Schema {
	return &t.schema
}

func (t Transaction) Scope() string {
	return "private"
}

func (t Transaction) RPC() Rpc {
	return t.r
}
