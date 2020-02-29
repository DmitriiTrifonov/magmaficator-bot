package main

import (
	"fmt"
	"github.com/DmitriiTrifonov/magmafier-bot/ctr"
	"github.com/DmitriiTrifonov/magmafier-bot/magmafier"
	tb "gopkg.in/tucnak/telebot.v2"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"net/http"
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

	message := "This is the Magmafier bot.\n" +
		"It uses block cipher \"Magma\" from GOST 34.12-2018.\n" +
		"It's only ciphering green and blue channels for entertainment purposes.\n" +
		"If you want to set a custom key please add a message to your photo\n" +
		"Have fun! This was a theme for my thesis."

	b.Handle("/start", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, message)
	})

	b.Handle("/help", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, message)
	})

	b.Handle("/about", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, message)
	})

	b.Handle(tb.OnDocument, func(m *tb.Message) {
		mgm := magmafier.Magma{}
		caption := m.Caption
		log.Println("Caption:", caption)
		key := magmafier.MakeKeyFromString(caption)
		log.Println("Key:", key)
		mgm.SetKey(key)
		mgm.SetSubKeys()
		photoUrl := m.Document.File.FileID
		url, err := b.FileURLByID(photoUrl)
		log.Println(url)
		resp, err := http.Get(url)
		log.Println(resp.Body)
		img, format, err := image.Decode(resp.Body)
		log.Println("image format is", format)
		err = resp.Body.Close()
		x := img.Bounds().Dx()
		y := img.Bounds().Dy()
		mod := image.NewRGBA(image.Rect(0, 0, x, y))
		counter := make([]byte, 0)
		counter = append(counter, key[24:32]...)
		for i := 0; i < x; i++ {
			for j := 0; j < y; j++ {
				col := img.At(i, j)
				r, g, b, _ := col.RGBA()

				r16 := uint16(r)
				g16 := uint16(g)
				b16 := uint16(b)

				rb := ui16tob(r16)
				gb := ui16tob(g16)
				bb := ui16tob(b16)

				block := make([]byte, 0)
				block = append(block, rb[0:1]...)
				block = append(block, gb[0:1]...)
				block = append(block, bb[0:1]...)

				cipher := ctr.CTRCrypt(block, counter, &mgm)

				mod.Set(i, j, color.RGBA{
					R: cipher[0],
					G: cipher[1],
					B: cipher[2],
					A: 255,
				})
			}
		}
		keyFile := fmt.Sprintf("%x", key)
		outFile, err := os.Create(keyFile + ".jpg")
		log.Println("File created:", keyFile+".jpg")

		err = png.Encode(outFile, mod)

		if err != nil {
			_, _ = b.Send(m.Sender, "Cannot process the photo")
		}

		p := &tb.Document{File: tb.FromDisk(keyFile + ".jpg"), Caption: keyFile, FileName: keyFile + ".jpg"}
		_, _ = b.Send(m.Sender, p)
		outFile.Close()
		os.Remove(keyFile + ".jpg")
		log.Println("File deleted:", keyFile+".jpg")

	})

	b.Start()

}

func ui16tob(a uint16) []byte {
	b := make([]byte, 2)
	b[1] = byte(a)
	b[0] = byte(a >> 8)
	return b
}

func btoui16(b []byte) (a uint16, err error) {
	a = uint16(b[0])
	a <<= 8
	a |= uint16(b[1])
	if len(b) > 2 {
		err = &ByteError{}
	}
	return
}

type ByteError struct {
}

func (b *ByteError) Error() string { return "Slice length is not correct" }
