package nvidia

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

const store = "https://www.nvidia.com/en-us/geforce/graphics-cards/30-series/rtx-3080/"
const api = "https://api-prod.nvidia.com/direct-sales-shop/DR/products/en_us/USD/%s"

type StoreAPI struct {
	CheckoutPage string
	ID           string
	ProductName  string
}

func (s *StoreAPI) Product() string {
	return s.ProductName
}

func (s *StoreAPI) Store() string {
	return "NVIDIA Store"
}

func (s *StoreAPI) URL() string {
	return s.CheckoutPage
}

func (s *StoreAPI) CheckAvailability(ctx context.Context, client *http.Client) (bool, error) {
	url := fmt.Sprintf(api, s.ID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Printf("nvidia-api: Failed to create request: %s", err)
		return false, err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("nvidia-api: Error executing HTTP request: %s", err)
		return false, err
	}

	defer resp.Body.Close()
	var storeResponse nvidiaStoreApiResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&storeResponse)
	if err != nil {
		log.Printf("nvidia-api: Error parsing json response: %s", err)
		return false, err
	}

	stockValue := storeResponse.Products.Product[0].InventoryStatus.ProductIsInStock
	inStock, err := strconv.ParseBool(stockValue)

	return inStock, err
}

type nvidiaStoreApiResponse struct {
	Products nvidiaStoreApiProducts `json:"products"`
}

type nvidiaStoreApiProducts struct {
	Product []nvidiaStoreApiProduct `json:"product"`
}

type nvidiaStoreApiProduct struct {
	ID              int64                         `json:"id"`
	Name            string                        `json:"name"`
	DisplayName     string                        `json:"displayName"`
	SKU             string                        `json:"sku"`
	InventoryStatus nvidiaStoreApiInventoryStatus `json:"inventoryStatus"`
}

type nvidiaStoreApiInventoryStatus struct {
	ProductIsInStock string `json:"productIsInStock"`
}
