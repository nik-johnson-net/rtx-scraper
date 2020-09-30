package bestbuy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const pageURLTemplate string = "https://www.bestbuy.com/api/tcfb/model.json?paths=%%5B%%5B%%22shop%%22%%2C%%22buttonstate%%22%%2C%%22v5%%22%%2C%%22item%%22%%2C%%22skus%%22%%2C%d%%2C%%22conditions%%22%%2C%%22NONE%%22%%2C%%22destinationZipCode%%22%%2C%s%%2C%%22storeId%%22%%2C%d%%2C%%22context%%22%%2C%%22cyp%%22%%2C%%22addAll%%22%%2C%%22false%%22%%5D%%5D&method=get"
const storeURL string = "https://www.bestbuy.com/site/%d.p?skuId=%d"

type Store struct {
	ProductName string
	SkuID       int
	Zip         string
	StoreID     int
}

func (s *Store) Product() string {
	return s.ProductName
}

func (s *Store) Store() string {
	return "Best Buy"
}

func (s *Store) URL() string {
	return fmt.Sprintf(storeURL, s.SkuID, s.SkuID)
}

func (s *Store) CheckAvailability(ctx context.Context, client *http.Client) (bool, error) {
	url := fmt.Sprintf(pageURLTemplate, s.SkuID, s.Zip, s.StoreID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("bestbuy: Failed to create request: %s", err)
		return false, err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("bestbuy: Error executing HTTP request: %s", err)
		return false, err
	}

	defer resp.Body.Close()
	var buffer bytes.Buffer
	var jsonResponse response

	decoder := json.NewDecoder(io.TeeReader(resp.Body, &buffer))
	err = decoder.Decode(&jsonResponse)
	if err != nil {
		log.Printf("bestbuy: Error parsing json response: %s", err)
		return false, nil
	}

	stringSku := fmt.Sprintf("%d", s.SkuID)
	sku, ok := jsonResponse.JsonGraph.Shop.Buttonstate.V5.Item.Skus[stringSku]
	if !ok {
		log.Println("bestbuy: response body:", buffer.String())
		log.Printf("bestbuy: missing sku in response %s", stringSku)
		return false, fmt.Errorf("invalid response")
	}

	zip, ok := sku.Conditions.NONE.DestinationZipCode[s.Zip]
	if !ok {
		log.Println("bestbuy: response body:", buffer.String())
		log.Printf("bestbuy: missing zip in response %s", s.Zip)
		return false, fmt.Errorf("invalid response")
	}

	store, ok := zip.StoreId[fmt.Sprintf("%d", s.StoreID)]
	if !ok {
		log.Println("bestbuy: response body:", buffer.String())
		log.Printf("bestbuy: missing store in response %d", s.StoreID)
		return false, fmt.Errorf("invalid response")
	}

	addAll, ok := store.Context.Cyp.AddAll["false"]
	if !ok {
		log.Println("bestbuy: response body:", buffer.String())
		log.Printf("bestbuy: missing addAll in response %s", "false")
		return false, fmt.Errorf("invalid response")
	}
	products := addAll.Value.ButtonStateResponseInfos

	for _, product := range products {
		if product.SkuId == stringSku {
			log.Printf("bestbuy: found %s (%d) state %s", s.ProductName, s.SkuID, product.ButtonState)
			return product.ButtonState == "ADD_TO_CART", nil
		}
	}

	return false, fmt.Errorf("sku not found")
}

type response struct {
	JsonGraph jsonGraph `json:"jsonGraph"`
}

type jsonGraph struct {
	Shop shop `json:"shop"`
}

type shop struct {
	Buttonstate buttonstate `json:"buttonstate"`
}

type buttonstate struct {
	V5 v5 `json:"v5"`
}

type v5 struct {
	Item item `json:"Item"`
}

type item struct {
	Skus map[string]skuItem `json:"skus"`
}

type skuItem struct {
	Conditions conditions `json:"conditions"`
}

type conditions struct {
	NONE none `json:"NONE"`
}

type none struct {
	DestinationZipCode map[string]zipEntry `json:"destinationZipCode"`
}

type zipEntry struct {
	StoreId map[string]storeEntry `json:"storeId"`
}

type storeEntry struct {
	Context storeContext `json:"context"`
}

type storeContext struct {
	Cyp cyp `json:"cyp"`
}

type cyp struct {
	AddAll map[string]addAllEntry `json:"addAll"`
}

type addAllEntry struct {
	Value value `json:"value"`
}

type value struct {
	ButtonStateResponseInfos []buttonStateResponseInfos `json:"buttonStateResponseInfos"`
}

type buttonStateResponseInfos struct {
	SkuId       string `json:"skuId"`
	ButtonState string `json:"buttonState"`
	DisplayText string `json:"displayText"`
}
