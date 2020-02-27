package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	publicURL := os.Getenv("PUBLIC_URL")
	token := os.Getenv("TOKEN")

	webhook := &tb.Webhook{
		Listen:   ":" + port,
		Endpoint: &tb.WebhookEndpoint{PublicURL: publicURL},
	}

	pref := tb.Settings{
		Token:  token,
		Poller: webhook,
	}

	b, err := tb.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	b.Handle("/start", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, "This is the Mamgafier bot.\n It uses block cipher \"Mamga\" from GOST  ")
	})

	b.Handle(tb.OnPhoto, func(m *tb.Message) {
		photo := m.Photo.File
		_, _ = b.Send(m.Sender, photo)
	})

	b.Start()

}
