package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/nik-johnson-net/rtx-scraper/notifiers"
)

type FakeStore struct {
	results []bool
}

func (f *FakeStore) Store() string {
	return "FakeStore"
}

func (f *FakeStore) Product() string {
	return "FakeProduct"
}

func (f *FakeStore) URL() string {
	return "https://localhost/"
}

type NotifyDetector struct {
	Notified bool
}

func (n *NotifyDetector) Notify(ctx context.Context, product, store, url string, instock bool) error {
	n.Notified = true
	return nil
}

func (f *FakeStore) CheckAvailability(ctx context.Context, client *http.Client) (bool, error) {
	r := f.results[0]
	f.results = f.results[1:len(f.results)]
	return r, nil
}

func TestPoll(t *testing.T) {
	testCases := [][]bool{
		{false, false},
		{false, true},
		{true, true},
		{true, false},
	}

	testResults := []bool{
		false,
		true,
		false,
		true,
	}

	for i, testCase := range testCases {
		if testResults[i] {
			t.Logf("Testing if %v results in notification", testCase)
		} else {
			t.Logf("Testing if %v results in no notification", testCase)
		}
		notify := &NotifyDetector{}
		pe := &PollEntry{
			store: &FakeStore{
				results: testCase,
			},
			notifier: []notifiers.Notifier{notify},
		}

		notified, err := pe.Poll(context.TODO(), &http.Client{}, false)
		if notified {
			t.Fatalf("Notified is true")
		}
		if err != nil {
			t.Fatalf("Err: %s", err)
		}

		notified, err = pe.Poll(context.TODO(), &http.Client{}, true)
		if notified != testResults[i] {
			t.Fatalf("Notified != %v", testResults[i])
		}
		if err != nil {
			t.Fatalf("Err: %s", err)
		}
		if notify.Notified != testResults[i] {
			t.Fatalf("Notifier was not notified")
		}
	}
}
