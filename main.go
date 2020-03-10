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
		"It ciphers all three channels using CTR mode for entertainment purposes.\n" +
		"To decipher a photo you should sent to bot a ciphered PNG file as a document.\n" +
		"To set a key simply write it to bot. It's a temporal operation.\n" +
		"After some time the key will be flushed.\n" +
		"If you want to set a custom key please add a caption to your photo\n" +
		"Have fun! This was a theme for my thesis."

	b.Handle("/start", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, message)
	})

	b.Handle("/help", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, message)
	})

	b.Handle("/about", func(fm *tb.Message) {
		_, _ = b.Send(fm.Sender, message)
	})

	b.Handle(tb.OnText, func(hm *tb.Message) {

		_, _ = b.Send(hm.Sender, "Key has been set")

		b.Handle(tb.OnDocument, func(m *tb.Message) {
			caption := m.Caption
			log.Println("Caption:", caption)
			if caption == "" {
				caption = hm.Text
			}
			key := magmafier.MakeKeyFromString(caption)
			log.Println("Key:", key)
			photoUrl := m.Document.File.FileID
			url, err := b.FileURLByID(photoUrl)
			if err != nil {
				_, _ = b.Send(m.Sender, "Cannot process the photo")
				return
			}
			process(url, caption, m, b)
		})
	})

	b.Handle(tb.OnDocument, func(m *tb.Message) {
		processDocument(m, b)
	})

	b.Handle(tb.OnPhoto, func(m *tb.Message) {
		convertPhoto(m, b)
	})

	b.Start()

}

func ui16tob(a uint16) []byte {
	b := make([]byte, 2)
	b[1] = byte(a)
	b[0] = byte(a >> 8)
	return b
}

func convertPhoto(m *tb.Message, b *tb.Bot) {

	caption := m.Caption
	log.Println("Caption:", caption)
	photoUrl := m.Photo.File.FileID
	url, err := b.FileURLByID(photoUrl)
	if err != nil {
		_, _ = b.Send(m.Sender, "Cannot process the photo")
		return
	}
	process(url, caption, m, b)

}

func processDocument(m *tb.Message, b *tb.Bot) {

	caption := m.Caption
	log.Println("Caption:", caption)
	photoUrl := m.Document.File.FileID
	url, err := b.FileURLByID(photoUrl)
	if err != nil {
		_, _ = b.Send(m.Sender, "Cannot process the photo")
		return
	}
	process(url, caption, m, b)
}

func process(url string, caption string, m *tb.Message, b *tb.Bot) {
	key := magmafier.MakeKeyFromString(caption)
	log.Println("Key:", key)
	mgm := magmafier.Magma{}
	mgm.SetKey(key)
	mgm.SetSubKeys()
	counter := make([]byte, 0)
	counter = append(counter, key[24:32]...)

	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		_, _ = b.Send(m.Sender, "Cannot process the photo")
		return
	}

	log.Println(resp.Body)
	img, format, err := image.Decode(resp.Body)
	if err != nil {
		_, _ = b.Send(m.Sender, "Cannot process the photo")
		return
	}
	log.Println("image format is", format)
	err = resp.Body.Close()

	x := img.Bounds().Dx()
	y := img.Bounds().Dy()
	mod := image.NewRGBA(image.Rect(0, 0, x, y))

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
	outFile, err := os.Create(keyFile + ".png")
	log.Println("File created:", keyFile+".png")

	err = png.Encode(outFile, mod)

	if err != nil {
		_, _ = b.Send(m.Sender, "Cannot process the photo")
	}

	p := &tb.Document{File: tb.FromDisk(keyFile + ".png"), Caption: keyFile, FileName: keyFile + ".png"}
	_, _ = b.Send(m.Sender, p)
	err = outFile.Close()
	err = os.Remove(keyFile + ".png")
	log.Println("File deleted:", keyFile+".png")
}
