package cmd

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/graphql-go/handler"
	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"wahyuade.com/simple-e-commerce/schemas"
)

var runGraphqlCmd = &cobra.Command{
	Use:   "graphql",
	Short: "Run the graphql server",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Graphql server is spinning on port 7000...")

		web := fiber.New()
		s := schemas.Construct()
		graphQL := web.Group("/graphql")
		graphQL.All("/user", routingSchema(s.User()))
		graphQL.All("/product", routingSchema(s.Product()))
		graphQL.All("/transaction", routingSchema(s.Transaction()))
		graphQL.All("/order", routingSchema(s.Order()))
		web.Listen("0.0.0.0:7000")
	},
}

func routingSchema(schema schemas.Schema) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		s := schema.Context(c)
		uuid := uuid.NewString()
		c.Locals("uuid", uuid)
		c.Set("correlation-id", uuid)
		headers := c.GetReqHeaders()
		authorization := headers["Authorization"]
		user := s.RPC().User().GetBySession(c.Context(), uuid, authorization)
		c.Locals("user", user)
		if s.Scope() == "private" {
			if authorization == "" {
				return unAuthorized(c)
			}
			if user.Uuid == "" {
				return unAuthorized(c)
			}
		}
		h := handler.New(&handler.Config{
			Schema: s.Config(),
			Pretty: true,
		})
		fasthttpadaptor.NewFastHTTPHandler(h)(c.Context())
		return nil
	}
}

func unAuthorized(c *fiber.Ctx) error {
	type Err struct {
		Message string `json:"message"`
	}
	return c.Status(403).JSON(struct {
		Data   interface{} `json:data`
		Errors []Err       `json:"errors"`
	}{
		Data: nil,
		Errors: []Err{
			Err{
				Message: "Silahkan login terlebih dahulu",
			},
		},
	})
}
