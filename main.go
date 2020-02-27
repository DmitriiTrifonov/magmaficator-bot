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

	message := "This is the Mamgafier bot.\n" +
		"It uses block cipher \"Mamga\" from GOST 34.12-2018.\n" +
		"It's only ciphering green and blue channels for entertainment purposes.\n" +
		"If you want to set a custom key please add a message to your photo" +
		"Have fun!"

	b.Handle("/start", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, message)
	})

	b.Handle("/help", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, message)
	})

	b.Handle("/about", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, message)
	})

	b.Handle(tb.OnPhoto, func(m *tb.Message) {
		photo := m.Photo
		_, _ = b.Send(m.Sender, photo)
	})

	b.Start()

}
