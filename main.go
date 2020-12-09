package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nik-johnson-net/rtx-scraper/notifiers"
	"github.com/nik-johnson-net/rtx-scraper/stores"
	"github.com/nik-johnson-net/rtx-scraper/stores/bestbuy"
	"github.com/nik-johnson-net/rtx-scraper/stores/nvidia"
)

const UserAgent string = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:80.0) Gecko/20100101 Firefox/80.0"

var tripper = &SetUserAgentTransport{
	UserAgent: UserAgent,
}

var client = &http.Client{
	Transport: tripper,
	Timeout:   10 * time.Second,
}

type PollEntry struct {
	store     stores.Store
	notifier  []notifiers.Notifier
	lastValue bool
}

func (p *PollEntry) Poll(ctx context.Context, client *http.Client, sendNotification bool) (bool, error) {
	log.Printf("Checking %s for %s availability", p.store.Store(), p.store.Product())
	instock, err := p.store.CheckAvailability(ctx, client)
	if err != nil {
		log.Printf("poll: failed to check availability: %s", err)
		return false, err
	}

	shouldNotify := sendNotification && (instock != p.lastValue)
	p.lastValue = instock

	if shouldNotify && (p.notifier != nil) {
		log.Printf("%s - %s availability changed to %v", p.store.Store(), p.store.Product(), instock)
		for _, notifier := range p.notifier {
			err = notifier.Notify(ctx, p.store.Product(), p.store.Store(), p.store.URL(), instock)
			if err != nil {
				log.Printf("poll: failed to notify: %s", err)
			}
		}
	}

	return shouldNotify, nil
}

type SetUserAgentTransport struct {
	UserAgent string
}

func (s *SetUserAgentTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", s.UserAgent)
	return http.DefaultTransport.RoundTrip(r)
}

func main() {
	notifiersList := []notifiers.Notifier{}

	notifyList2 := []notifiers.Notifier{}

	pollers := []*PollEntry{
		{
			store: &bestbuy.Store{
				ProductName: "RTX 3080",
				SkuID:       6429440,
				Zip:         "80020",
				StoreID:     186,
			},
			notifier: notifiersList,
		},
		{
			store: &nvidia.Store{
				ProductName:  "RTX 3080",
				ProductSKU:   30042,
				SearchString: "RTX%203080",
			},
			notifier: notifiersList,
		},
		{
			store: &nvidia.StoreAPI{
				CheckoutPage: "https://www.nvidia.com/en-us/geforce/graphics-cards/30-series/rtx-3080/",
				ID:           "5438481700",
				ProductName:  "RTX 3080",
			},
			notifier: notifiersList,
		},
		// Xbox for Someone Else
		{
			store: &bestbuy.Store{
				ProductName: "Xbox Series X",
				SkuID:       6428324,
				Zip:         "80020",
				StoreID:     186,
			},
			notifier: notifyList2,
		},
	}

	// Flag parsing

	testNotifierFlag := flag.Int("test-notifier", 0, fmt.Sprintf("Which notifier to send [0..%d]", len(notifiersList)-1))
	testPollerFlag := flag.Int("test-poller", 0, fmt.Sprintf("Which poller to test [0..%d]", len(pollers)-1))

	flag.Parse()

	if flagSet("test-notifier") {
		err := testNotifier(notifiersList[*testNotifierFlag])
		if err != nil {
			log.Printf("Testing of notifier %d failed: %s\n", *testNotifierFlag, err)
			os.Exit(1)
		}
		log.Printf("Successfully tested notifier %d\n", *testNotifierFlag)
		os.Exit(0)
	}

	if flagSet("test-poller") {
		err := testPoller(pollers[*testPollerFlag])
		if err != nil {
			log.Printf("Testing of poller %d failed: %s\n", *testPollerFlag, err)
			os.Exit(1)
		}
		log.Printf("Successfully tested poller %d\n", *testPollerFlag)
		os.Exit(0)
	}

	// Main

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

func flagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func testNotifier(notifier notifiers.Notifier) error {
	if err := notifier.Notify(context.Background(), "TestProduct", "TestStore", "https://this-is-a-test/", true); err != nil {
		return err
	}
	if err := notifier.Notify(context.Background(), "TestProduct", "TestStore", "https://this-is-a-test/", false); err != nil {
		return err
	}
	return nil
}

func testPoller(poller *PollEntry) error {
	result, err := poller.Poll(context.Background(), client, false)
	if err == nil {
		log.Printf("Poller returned %v\n", result)
	}
	return err
}
