package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"wahyuade.com/simple-e-commerce/schemas"
)

type grapqlErrorModel struct {
	Message string `json:"message"`
}

type graphqlResponseModel struct {
	Data   interface{}        `json:"data"`
	Errors []grapqlErrorModel `json:"errors,omitempty"`
}

type graphqlQueryModel struct {
	Query string `json:"query"`
}

func getUrl(t *testing.T, module string) string {
	args := os.Args
	baseUrl := args[len(args)-1]
	u, err := url.Parse(baseUrl)
	require.Equal(t, nil, err)

	require.Equal(t, "http", u.Scheme, "BASE_URL must start with http://")
	require.Equal(t, "/graphql/", u.Path, "BASE_URL must have path: /graphql")

	require.Equal(t, checkModule(module), true, "Invalid module")
	return baseUrl + module
}

func checkModule(module string) bool {
	for _, m := range []string{"user", "product", "transaction", "order"} {
		if module == m {
			return true
		}
	}
	return false
}

func getTestLoginSession(t *testing.T) string {
	email := "test@grpc_test.com"
	password := "grpc_test"
	url := getUrl(t, "user")
	query := graphqlQueryModel{
		Query: fmt.Sprintf(`mutation {
			login(email: "%s", password: "%s") {
				session
			}
		}`, email, password),
	}

	body, _ := json.Marshal(query)
	res, err := http.Post(url, "application/json", bytes.NewBuffer(body))

	require.Equal(t, nil, err, "Make sure this test can access the server")
	require.Equal(t, 200, res.StatusCode)

	type loginResponse struct {
		Login schemas.User `json:"login"`
	}

	type mockGraphqlResponse struct {
		Data   loginResponse      `json:"data"`
		Errors []grapqlErrorModel `json:"errors"`
	}
	var response mockGraphqlResponse
	json.NewDecoder(res.Body).Decode(&response)
	require.Equal(t, 0, len(response.Errors))
	return response.Data.Login.Session
}
