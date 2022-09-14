package schemas

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/graphql-go/graphql"
	"wahyuade.com/simple-e-commerce/helpers"
)

type User struct {
	path   string
	schema graphql.Schema
	r      Rpc
	ctx    *fiber.Ctx

	Uuid     string `json:"uuid,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	Session  string `json:"session,omitempty"`
}

func userType(name string, field []string) *graphql.Object {
	allField := map[string]*graphql.Field{
		"uuid":     typeString,
		"name":     typeString,
		"email":    typeString,
		"password": typeString,
		"session":  typeString,
	}
	allowedField := getAllowedField(allField, field)
	return graphql.NewObject(
		graphql.ObjectConfig{
			Name:   name,
			Fields: *allowedField,
		},
	)
}

func (u User) userQuery() *graphql.Field {
	return &graphql.Field{
		Type: userType("user", []string{"name", "email", "uuid", "session"}),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			user := u.ctx.Locals("user")
			if user == nil {
				return nil, errors.New("Silahkan login terlebih dahulu")
			}
			u := user.(User)
			if u.Uuid == "" {
				return nil, errors.New("Silahkan login terlebih dahulu")
			}
			return u, nil
		},
	}
}

func (u User) loginMutation() *graphql.Field {
	return &graphql.Field{
		Type:        userType("login", []string{"name", "email", "uuid", "session"}),
		Description: "Login untuk user",
		Args: graphql.FieldConfigArgument{
			"email": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"password": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			email := p.Args["email"].(string)
			password := p.Args["password"].(string)
			correlationId := u.ctx.Locals("uuid").(string)
			status, result := u.r.User().GetByEmail(p.Context, correlationId, email)
			if !status {
				return nil, errors.New("user not found")
			}
			isPasswordMatch, _ := helpers.MatchHash(password, result.Password)
			if !isPasswordMatch {
				return nil, errors.New("invalid password")
			}
			session := helpers.RandomString(64, correlationId)
			result.Session = session
			setSession := u.r.User().SaveSession(p.Context, correlationId, result)
			if setSession {
				return result, nil
			}
			return nil, errors.New("Internal server error")
		},
	}
}

func (u User) registerMutation() *graphql.Field {
	return &graphql.Field{
		Type:        userType("register", []string{"name", "email", "uuid"}),
		Description: "Pendaftaran user",
		Args: graphql.FieldConfigArgument{
			"name": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"email": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"password": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			correlationId := u.ctx.Locals("uuid").(string)
			password, _ := helpers.Hash(p.Args["password"].(string))
			user := User{
				Name:     p.Args["name"].(string),
				Email:    p.Args["email"].(string),
				Password: password,
			}
			status, result := u.r.User().Register(p.Context, correlationId, user)
			if status {
				return result, nil
			}
			return nil, errors.New("User already registered")
		},
	}
}

func newUser(r Rpc) User {
	user := User{r: r}
	return user
}

func (u User) Path() string {
	return u.path
}

func (u User) Config() *graphql.Schema {
	return &u.schema
}

func (u User) Scope() string {
	return "public"
}

func (u User) Context(c *fiber.Ctx) Schema {
	user := User{r: u.r, ctx: c}
	query := graphql.NewObject(graphql.ObjectConfig{
		Name: "UserQuery",
		Fields: graphql.Fields{
			"user": user.userQuery(),
		},
	})

	mutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "UserMutation",
		Fields: graphql.Fields{
			"login":    user.loginMutation(),
			"register": user.registerMutation(),
		},
	})
	s, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    query,
		Mutation: mutation,
	})
	if err != nil {
		log.Fatalf("Gagal membuat schema, error: %v", err.Error())
	}
	user.schema = s
	return user
}

func (u User) RPC() Rpc {
	return u.r
}
