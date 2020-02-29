package main

import (
	"fmt"
	"github.com/DmitriiTrifonov/magmafier-bot/ctr"
	"github.com/DmitriiTrifonov/magmafier-bot/magmafier"
	tb "gopkg.in/tucnak/telebot.v2"
	"image"
	"image/color"
	"image/jpeg"
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

	b.Handle(tb.OnPhoto, func(m *tb.Message) {
		mgm := magmafier.Magma{}
		caption := m.Caption
		log.Println("Caption:", caption)
		key := magmafier.MakeKeyFromString(caption)
		log.Println("Key:", key)
		mgm.SetKey(key)
		photoUrl := m.Photo.File.FileID
		url, err := b.FileURLByID(photoUrl)
		log.Println(url)
		resp, err := http.Get(url)
		log.Println(resp.Body)
		img, err := jpeg.Decode(resp.Body)
		err = resp.Body.Close()
		x := img.Bounds().Dx()
		y := img.Bounds().Dy()
		mod := image.NewRGBA(image.Rect(0, 0, x, y))
		counter := make([]byte, 8)
		copy(counter, ctr.Vector)
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
				block = append(block, rb...)
				block = append(block, gb...)
				block = append(block, bb...)

				//cipher := ctr.CTRCrypt(block, counter, &mgm)

				newR16, _ := btoui16(block[0:2])
				newG16, _ := btoui16(block[2:4])
				newB16, _ := btoui16(block[4:6])

				mod.Set(i, j, color.RGBA64{
					R: newR16,
					G: newG16,
					B: newB16,
					A: 65535,
				})
			}
		}
		keyFile := fmt.Sprintf("%x", key)
		outFile, err := os.Create(keyFile + ".jpg")
		log.Println("File created:", keyFile+".jpg")

		err = jpeg.Encode(outFile, mod, nil)

		if err != nil {
			_, _ = b.Send(m.Sender, "Cannot process the photo")
		}

		p := &tb.Photo{File: tb.FromDisk(keyFile + ".jpg"), Caption: keyFile}
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
