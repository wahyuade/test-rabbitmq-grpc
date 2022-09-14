package schemas

import (
	"github.com/gofiber/fiber/v2"
	"github.com/graphql-go/graphql"
)

var typeString = &graphql.Field{
	Type: graphql.String,
}

var typeInt = &graphql.Field{
	Type: graphql.Int,
}

var typeDatetime = &graphql.Field{
	Type: graphql.DateTime,
}

type Schema interface {
	Path() string
	Config() *graphql.Schema
	Scope() string
	Context(*fiber.Ctx) Schema
	RPC() Rpc
}

type ISchema interface {
	User() Schema
	Product() Schema
	Transaction() Schema
	Order() Schema
}

type schemas struct {
	rpc Rpc
}

func (s schemas) User() Schema {
	return newUser(s.rpc)
}

func (s schemas) Order() Schema {
	return newOrder(s.rpc)
}
func (s schemas) Product() Schema {
	return newProduct(s.rpc)
}
func (s schemas) Transaction() Schema {
	return newTransaction(s.rpc)
}

func Construct() ISchema {
	rpc := InitRpc()
	return schemas{
		rpc: rpc,
	}
}

func getAllowedField(allField map[string]*graphql.Field, allowedField []string) *graphql.Fields {
	fields := graphql.Fields{}
	if len(allowedField) == 0 {
		fields = allField
	} else {
		for _, f := range allowedField {
			if allField[f].Type != nil {
				fields[f] = allField[f]
			}
		}
	}
	return &fields
}
