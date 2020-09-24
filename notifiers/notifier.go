package notifiers

import "context"

type Notifier interface {
	Notify(ctx context.Context, product string, store string, url string, instock bool) error
}
