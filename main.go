package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/nik-johnson-net/rtx-scraper/notifiers"
	"github.com/nik-johnson-net/rtx-scraper/notifiers/smtp"
	"github.com/nik-johnson-net/rtx-scraper/stores"
	"github.com/nik-johnson-net/rtx-scraper/stores/bestbuy"
	"github.com/nik-johnson-net/rtx-scraper/stores/nvidia"
)

const UserAgent string = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:80.0) Gecko/20100101 Firefox/80.0"

type PollEntry struct {
	store     stores.Store
	notifier  []notifiers.Notifier
	lastValue bool
}

func (p *PollEntry) Poll(ctx context.Context, client *http.Client, sendNotification bool) {
	log.Printf("Checking %s for %s availability", p.store.Store(), p.store.Product())
	instock, err := p.store.CheckAvailability(ctx, client)
	if err != nil {
		log.Printf("poll: failed to check availability: %s", err)
		return
	}

	if sendNotification && instock != p.lastValue && p.notifier != nil {
		log.Printf("%s - %s availability changed to %v", p.store.Store(), p.store.Product(), instock)
		for _, notifier := range p.notifier {
			err = notifier.Notify(ctx, p.store.Product(), p.store.Store(), p.store.URL(), instock)
			if err != nil {
				log.Printf("poll: failed to notify: %s", err)
				return
			}
		}
	}
	p.lastValue = instock
}

type SetUserAgentTransport struct {
	UserAgent string
}

func (s *SetUserAgentTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", s.UserAgent)
	return http.DefaultTransport.RoundTrip(r)
}

func main() {
	notifiers := []notifiers.Notifier{
		// Nik (Phone)
		&smtp.SMTP{
			Server: "smtp.echo1.jnstw.net:25",
			From:   "rtx-scraper@jnstw.net",
			To:     "9793938339@txt.att.net",
		},
		// Nik
		&smtp.SMTP{
			Server: "smtp.echo1.jnstw.net:25",
			From:   "rtx-scraper@jnstw.net",
			To:     "nik@nikjohnson.net",
		},
		// Joe (Gunsjoe)
		&smtp.SMTP{
			Server: "smtp.echo1.jnstw.net:25",
			From:   "rtx-scraper@jnstw.net",
			To:     "joe@indra.com",
		},
		// TuggerNutts
		&smtp.SMTP{
			Server: "smtp.echo1.jnstw.net:25",
			From:   "rtx-scraper@jnstw.net",
			To:     "seak789@yahoo.com",
		},
		// Emerson
		&smtp.SMTP{
			Server: "smtp.echo1.jnstw.net:25",
			From:   "rtx-scraper@jnstw.net",
			To:     "squidboy54@me.com",
		},
	}

	pollers := []PollEntry{
		{
			store: &bestbuy.Store{
				ProductName: "RTX 3080",
				SkuID:       6429440,
				Zip:         "80020",
				StoreID:     186,
			},
			notifier: notifiers,
		},
		{
			store: &nvidia.Store{
				ProductName:  "RTX 3080",
				ProductSKU:   30042,
				SearchString: "RTX%203080",
			},
			notifier: notifiers,
		},
	}

	tripper := &SetUserAgentTransport{
		UserAgent: UserAgent,
	}

	client := &http.Client{
		Transport: tripper,
		Timeout:   5 * time.Second,
	}

	for _, store := range pollers {
		store.Poll(context.Background(), client, false)
	}

	time.Sleep(10 * time.Second)

	for {
		ctx := context.Background()

		for _, store := range pollers {
			store.Poll(ctx, client, true)
		}

		time.Sleep(10 * time.Second)
	}
}
