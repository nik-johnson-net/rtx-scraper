package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Discord struct {
	ID     string
	Token  string
	Client *http.Client
}

func (d *Discord) url() string {
	return fmt.Sprintf("https://discordapp.com/api/webhooks/%s/%s", d.ID, d.Token)
}

func (d *Discord) Notify(ctx context.Context, product string, store string, url string, instock bool) error {
	var message string
	if instock {
		message = fmt.Sprintf("%s is now in stock at %s: %s", product, store, url)
	} else {
		message = fmt.Sprintf("%s is now out of stock at %s: %s", product, store, url)
	}

	data := discordWebhook{
		Content: message,
		AllowedMentions: allowedMentions{
			Parse: []string{},
		},
	}

	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}

	resp, err := d.Client.Post(d.url(), "application/json", &body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("discord: error reading body:", err)
		}
		return fmt.Errorf("discord API returned %d: %s", resp.StatusCode, body)
	}

	return nil
}

type discordWebhook struct {
	Content         string          `json:"content"`
	Username        string          `json:"username"`
	AvatarURL       string          `json:"avatar_url"`
	TTS             bool            `json:"tts"`
	Embeds          []embed         `json:"embeds"`
	AllowedMentions allowedMentions `json:"allowed_mentions"`
}

type allowedMentions struct {
	Parse []string `json:"parse"`
	Users []string `json:"users"`
	Roles []string `json:"roles"`
}

type embed struct {
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Timestamp   string `json:"timestamp"`
	Color       int    `json:"color"`
}
