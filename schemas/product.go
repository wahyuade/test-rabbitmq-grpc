package schemas

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/graphql-go/graphql"
)

type Product struct {
	path   string
	schema graphql.Schema
	r      Rpc
	ctx    *fiber.Ctx

	Uuid        string `json:"uuid,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Stock       int    `json:"stock,omitempty"`
	Price       int    `json:"price,omitempty"`
}

func productType(name string, field []string) *graphql.Object {
	allField := map[string]*graphql.Field{
		"uuid":        typeString,
		"name":        typeString,
		"description": typeString,
		"stock":       typeInt,
		"price":       typeInt,
	}
	allowedField := getAllowedField(allField, field)
	return graphql.NewObject(
		graphql.ObjectConfig{
			Name:   name,
			Fields: *allowedField,
		},
	)
}

func (pR Product) productsQuery() *graphql.Field {
	return &graphql.Field{
		Description: "Lihat semua produk yang ada",
		Type:        graphql.NewList(productType("products", []string{"uuid", "name", "price"})),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return pR.r.Product().ListProduct(p.Context, pR.ctx.Locals("uuid").(string)), nil
		},
	}
}

func (pR Product) productQuery() *graphql.Field {
	return &graphql.Field{
		Description: "Lihat produk secara detail",
		Type:        productType("product", []string{}),
		Args: graphql.FieldConfigArgument{
			"uuid": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			status, product := pR.r.Product().DetailProduct(p.Context, pR.ctx.Locals("uuid").(string), p.Args["uuid"].(string))
			if status {
				return product, nil
			}
			return nil, errors.New("produk tidak ditemukan")
		},
	}
}

func (pR Product) createProductMutation() *graphql.Field {
	return &graphql.Field{
		Description: "Membuat produk baru",
		Type:        productType("createProduct", []string{}),
		Args: graphql.FieldConfigArgument{
			"name": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"description": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"stock": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"price": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			product := Product{
				Name:        p.Args["name"].(string),
				Description: p.Args["name"].(string),
				Stock:       p.Args["stock"].(int),
				Price:       p.Args["price"].(int),
			}
			status, product := pR.r.Product().CreateProduct(
				p.Context,
				pR.ctx.Locals("uuid").(string),
				product,
			)
			if status {
				return product, nil
			}
			return nil, errors.New("internal server error")
		},
	}
}

func newProduct(rpc Rpc) Product {
	product := Product{
		r: rpc,
	}
	return product
}

func (p Product) Context(c *fiber.Ctx) Schema {
	product := Product{
		r:   p.r,
		ctx: c,
	}
	query := graphql.NewObject(graphql.ObjectConfig{
		Name: "ProductQuery",
		Fields: graphql.Fields{
			"product":  product.productQuery(),
			"products": product.productsQuery(),
		},
	})
	mutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "ProductMutation",
		Fields: graphql.Fields{
			"createProduct": product.createProductMutation(),
		},
	})
	s, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    query,
		Mutation: mutation,
	})
	if err != nil {
		log.Fatalf("Gagal membuat schema, error: %v", err.Error())
	}
	product.schema = s
	return product
}

func (p Product) Path() string {
	return p.path
}

func (p Product) Config() *graphql.Schema {
	return &p.schema
}

func (p Product) Scope() string {
	return "private"
}

func (p Product) RPC() Rpc {
	return p.r
}
