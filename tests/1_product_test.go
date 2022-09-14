package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"wahyuade.com/simple-e-commerce/schemas"
)

func TestProductMutation_CreateProduct(t *testing.T) {
	url := getUrl(t, "product")
	session := getTestLoginSession(t)
	require.NotEqual(t, "", session)

	query := graphqlQueryModel{
		Query: `mutation {
			createProduct(
				name: "test produk"
				description: "contoh deskripsi"
				stock: 100
				price: 10000
			) {
				uuid
				name
				description
				stock
				price
			}
		}`,
	}
	body, _ := json.Marshal(query)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", session))

	resp, _ := client.Do(req)

	type mockCreateProduct struct {
		CreateProduct schemas.Product `json:"createProduct"`
	}

	type mockResponseCreateProduct struct {
		Data   mockCreateProduct  `json:"data"`
		Errors []grapqlErrorModel `json:"errors"`
	}

	var response mockResponseCreateProduct
	json.NewDecoder(resp.Body).Decode(&response)
	require.Equal(t, 0, len(response.Errors))

	require.NotEqual(t, "", response.Data.CreateProduct.Uuid)
}

func TestProductQuery_List(t *testing.T) {
	url := getUrl(t, "product")
	session := getTestLoginSession(t)
	require.NotEqual(t, "", session)

	query := graphqlQueryModel{
		Query: `{
			products {
				name,
				uuid,
				price
			}
		}`,
	}
	body, _ := json.Marshal(query)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", session))

	resp, _ := client.Do(req)

	type mockListProduct struct {
		Products []schemas.Product `json:"products"`
	}
	type mockResponseListProduct struct {
		Data   mockListProduct    `json:"data"`
		Errors []grapqlErrorModel `json:"errors"`
	}

	var response mockResponseListProduct
	json.NewDecoder(resp.Body).Decode(&response)

	require.Greater(t, len(response.Data.Products), 0)
}

func TestProductQuery_Detail(t *testing.T) {
	url := getUrl(t, "product")
	session := getTestLoginSession(t)
	require.NotEqual(t, "", session)

	query := graphqlQueryModel{
		Query: `{
			products {
				name,
				uuid,
				price
			}
		}`,
	}
	body, _ := json.Marshal(query)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", session))

	resp, _ := client.Do(req)

	type mockListProduct struct {
		Products []schemas.Product `json:"products"`
	}
	type mockResponseListProduct struct {
		Data   mockListProduct    `json:"data"`
		Errors []grapqlErrorModel `json:"errors"`
	}

	var response mockResponseListProduct
	json.NewDecoder(resp.Body).Decode(&response)

	require.Greater(t, len(response.Data.Products), 0)

	query = graphqlQueryModel{
		Query: fmt.Sprintf(`{
			product(uuid: "%s") {
				uuid,
				name,
				description,
				stock,
				price
			}
		}`, response.Data.Products[0].Uuid),
	}

	body, _ = json.Marshal(query)
	client = &http.Client{}
	req, _ = http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", session))
	resp, _ = client.Do(req)

	type mockProduct struct {
		Product schemas.Product `json:"product"`
	}

	type mockResponseDetailProduct struct {
		Data   mockProduct        `json:"data"`
		Errors []grapqlErrorModel `json:"errors"`
	}

	var responseDetail mockResponseDetailProduct
	json.NewDecoder(resp.Body).Decode(&responseDetail)
	require.Equal(t, 0, len(response.Errors))

	require.Equal(t, response.Data.Products[0].Uuid, responseDetail.Data.Product.Uuid)
}
