package smtp

import (
	"context"
	"fmt"
	"net/smtp"
)

type SMTP struct {
	Server string
	From   string
	To     string
}

func (s *SMTP) Notify(ctx context.Context, product string, store string, url string, instock bool) error {
	var msg string
	if instock {
		msg = fmt.Sprintf("%s in stock at %s - %s", product, store, url)
	} else {
		msg = fmt.Sprintf("%s no longer in stock at %s - %s", product, store, url)
	}
	return smtp.SendMail(s.Server, nil, s.From, []string{s.To}, []byte(msg))
}
