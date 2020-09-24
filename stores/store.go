package stores

import (
	"context"
	"net/http"
)

type Store interface {
	Product() string
	Store() string
	URL() string
	CheckAvailability(ctx context.Context, client *http.Client) (bool, error)
}
