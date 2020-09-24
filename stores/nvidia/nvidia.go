package nvidia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const searchURL string = "https://api.nvidia.partners/edge/product/search?page=1&limit=9&locale=en-us&search=%s&manufacturer=NVIDIA&manufacturer_filter=NVIDIA~1,ACER~0,ALIENWARE~0,ASUS~0,DELL~0,EVGA~2,GIGABYTE~2,HP~0,LENOVO~0,LG~0,MSI~3,PNY~0,RAZER~0,ZOTAC~0"
const storeURL string = "https://www.nvidia.com/en-us/shop/geforce/?page=1&limit=9&locale=en-us&manufacturer=NVIDIA&gpu=%s&gpu_filter=RTX%%203090~1,RTX%%203080~1,RTX%%203070~1,RTX%%202080%%20Ti~0,RTX%%202080%%20SUPER~0,RTX%%202080~0,RTX%%202070%%20SUPER~0,RTX%%202070~0,RTX%%202060%%20SUPER~1,RTX%%202060~0,GTX%%201660%%20Ti~0,GTX%%201660%%20SUPER~0,GTX%%201660~0,GTX%%201650%%20Ti~0,GTX%%201650%%20SUPER~0,GTX%%201650~0"

type Store struct {
	SearchString string
	ProductSKU   int
	ProductName  string
}

func (s *Store) Product() string {
	return s.ProductName
}

func (s *Store) Store() string {
	return "NVIDIA Store"
}

func (s *Store) URL() string {
	return fmt.Sprintf(storeURL, s.SearchString)
}

func (s *Store) CheckAvailability(ctx context.Context, client *http.Client) (bool, error) {
	url := fmt.Sprintf(searchURL, s.SearchString)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("nvidia: Failed to create request: %s", err)
		return false, err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("nvidia: Error executing HTTP request: %s", err)
		return false, err
	}

	defer resp.Body.Close()
	var jsonResponse searchResponse

	var buffer bytes.Buffer
	decoder := json.NewDecoder(io.TeeReader(resp.Body, &buffer))
	err = decoder.Decode(&jsonResponse)
	if err != nil {
		log.Printf("nvidia: Error parsing json response: %s", err)
		return false, nil
	}

	// log.Println("nvidia: response body:", buffer.String())

	for _, product := range jsonResponse.SearchedProducts.ProductDetails {
		if product.ProductID == s.ProductSKU {
			log.Printf("nvidia: found %s (%d) state %s", s.ProductName, s.ProductSKU, product.ProdStatus)
			return product.ProdStatus == "buy_now", nil
		}
	}

	return false, fmt.Errorf("sku not found")
}

type searchResponse struct {
	SearchedProducts searchedProducts `json:"searchedProducts"`
}

type searchedProducts struct {
	TotalProducts  int             `json:"totalProducts"`
	ProductDetails []productDetail `json:"productDetails"`
}

type productDetail struct {
	DisplayName    string `json:"displayName"`
	TotalCount     int    `json:"totalCount"`
	ProductID      int    `json:"productID"`
	DigitalRiverID string `json:"DigitalRiverID"`
	ProdStatus     string `json:"prdStatus"`
}
