package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"wahyuade.com/simple-e-commerce/helpers"
	"wahyuade.com/simple-e-commerce/schemas"
)

func TestUserMutation_CreateTestUser(t *testing.T) {
	email := "test@grpc_test.com"
	name := "wahyu"
	url := getUrl(t, "user")
	query := graphqlQueryModel{
		Query: fmt.Sprintf(`mutation {
			register(email: "%s", name: "%s", password: "grpc_test") {
				email
				name
			}
		}`, email, name),
	}
	body, _ := json.Marshal(query)
	res, err := http.Post(url, "application/json", bytes.NewBuffer(body))

	require.Equal(t, nil, err, "Make sure this test can access the server")
	require.Equal(t, 200, res.StatusCode)

	type registerResponse struct {
		Register schemas.User `json:"register"`
	}

	expectedModel := graphqlResponseModel{
		Data: registerResponse{
			Register: schemas.User{
				Email: email,
				Name:  name,
			},
		},
	}
	expectedJson, _ := json.Marshal(expectedModel)
	var response graphqlResponseModel
	json.NewDecoder(res.Body).Decode(&response)

	if response.Errors != nil {
		require.Equal(t, "User already registered", response.Errors[0].Message)
	} else {
		responseJson, _ := json.Marshal(response)
		require.JSONEq(t, string(expectedJson), string(responseJson))
	}
}

func TestUserMutation_Login(t *testing.T) {
	email := "test@grpc_test.com"
	url := getUrl(t, "user")
	query := graphqlQueryModel{
		Query: fmt.Sprintf(`mutation {
			login(email: "%s", password: "grpc_test") {
				email
			}
		}`, email),
	}
	body, _ := json.Marshal(query)
	res, err := http.Post(url, "application/json", bytes.NewBuffer(body))

	require.Equal(t, nil, err, "Make sure this test can access the server")
	require.Equal(t, 200, res.StatusCode)

	type loginResponse struct {
		Login schemas.User `json:"login"`
	}

	expected := graphqlResponseModel{
		Data: loginResponse{
			Login: schemas.User{
				Email: email,
			},
		},
	}
	expectedJson, _ := json.Marshal(expected)
	var response map[string]interface{}
	json.NewDecoder(res.Body).Decode(&response)

	responseJson, _ := json.Marshal(response)
	require.Equal(t, string(expectedJson), string(responseJson))
}

func TestUserMutation_Register(t *testing.T) {
	random := helpers.RandomString(5, uuid.NewString())
	email := fmt.Sprintf("%s@grpc_test.com", random)
	name := "wahyu"
	password := "grpc_test"
	url := getUrl(t, "user")
	query := graphqlQueryModel{
		Query: fmt.Sprintf(`mutation {
			register(email: "%s", name: "%s", password: "%s") {
				email
				name
			}
		}`, email, name, password),
	}
	body, _ := json.Marshal(query)
	res, err := http.Post(url, "application/json", bytes.NewBuffer(body))

	require.Equal(t, nil, err, "Make sure this test can access the server")
	require.Equal(t, 200, res.StatusCode)

	type registerResponse struct {
		Register schemas.User `json:"register"`
	}

	expectedModel := graphqlResponseModel{
		Data: registerResponse{
			Register: schemas.User{
				Email: email,
				Name:  name,
			},
		},
	}
	expectedJson, _ := json.Marshal(expectedModel)
	var response map[string]interface{}
	json.NewDecoder(res.Body).Decode(&response)
	responseJson, _ := json.Marshal(response)

	require.JSONEq(t, string(expectedJson), string(responseJson))
}

func TestUserQuery_User(t *testing.T) {
	// login
	email := "test@grpc_test.com"
	url := getUrl(t, "user")
	query := graphqlQueryModel{
		Query: fmt.Sprintf(`mutation {
			login(email: "%s", password: "grpc_test") {
				session
			}
		}`, email),
	}
	body, _ := json.Marshal(query)
	res, err := http.Post(url, "application/json", bytes.NewBuffer(body))

	require.Equal(t, nil, err, "Make sure this test can access the server")
	require.Equal(t, 200, res.StatusCode)

	type MockLoginModel struct {
		Login schemas.User `json:"login,omitempty"`
	}

	type MockUserModel struct {
		User schemas.User `json:"user,omitempty"`
	}

	type grapqlUserLogin struct {
		Data   MockLoginModel     `json:"data,omitempty"`
		Errors []grapqlErrorModel `json:"errors,omitempty"`
	}

	type grapqlUser struct {
		Data   MockUserModel      `json:"data,omitempty"`
		Errors []grapqlErrorModel `json:"errors,omitempty"`
	}

	response := grapqlUserLogin{}

	json.NewDecoder(res.Body).Decode(&response)
	require.Equal(t, 0, len(response.Errors), response.Errors)
	query = graphqlQueryModel{
		Query: `{
			user {
				session
			}
		}`,
	}
	client := &http.Client{}
	body, _ = json.Marshal(query)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", response.Data.Login.Session))

	resp, err := client.Do(req)
	responseUser := grapqlUser{}
	json.NewDecoder(resp.Body).Decode(&responseUser)

	require.Equal(t, response.Data.Login.Session, responseUser.Data.User.Session, err)
}
