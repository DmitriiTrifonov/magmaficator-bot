package main

import (
	"github.com/DmitriiTrifonov/magmafier-bot/magmafier"
	tb "gopkg.in/tucnak/telebot.v2"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
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
		for i := 0; i < x; i++ {
			for j := 0; j < y; j++ {
				c := img.At(i, j)
				r, g, b, a := c.RGBA()

				var rArr [4]byte
				convertToArray(r, &rArr)
				var gArr [4]byte
				convertToArray(g, &gArr)
				
				rgCipher := make([]byte, 0)
				rgCipher = append(rgCipher, rArr[:]...)
				rgCipher = append(rgCipher, gArr[:]...)


				var bArr [4]byte
				convertToArray(b, &bArr)
				var aArr [4]byte
				convertToArray(a, &aArr)
				
				baCipher := make([]byte, 0)
				baCipher = append(baCipher, bArr[:]...)
				baCipher = append(baCipher, aArr[:]...)

				rgCipher = mgm.Encrypt(rgCipher)
				baCipher = mgm.Encrypt(baCipher)

				var rgArr, baArr [8]byte
				copy(rgArr[:], rgCipher)
				copy(baArr[:], baCipher)
				_, newG := convertToUInt32(&rgArr)
				newB, _ := convertToUInt32(&baArr)

				mod.Set(i, j, color.RGBA64{
					R: uint16(r),
					G: uint16(newG * g),
					B: uint16(newB * b),
					A: 65535,
				})
			}
		}
		outFile, err := os.Create("changed.jpg")

		defer outFile.Close()
		png.Encode(outFile, mod)

		if err != nil {
			_, _ = b.Send(m.Sender, "Cannot process the photo")
		}

		p := &tb.Photo{File: tb.FromDisk("changed.jpg")}
		_, _ = b.Send(m.Sender, p)
	})

	b.Start()

}

func convertToUInt32(a *[8]byte) (uint32, uint32) {
	var r, r2 uint32
	for i := 0; i < 3; i++ {
		r |= uint32(a[i])
		r <<= 8
	}
	r |= uint32(a[3])
	for i := 4; i < 7; i++ {
		r2 |= uint32(a[i])
		r2 <<= 8
	}
	r2 |= uint32(a[7])
	return r, r2
}

func convertToArray(a uint32, arr *[4]byte) {
	arr[3] = byte(a)
	arr[2] = byte(a >> 8)
	arr[1] = byte(a >> 16)
	arr[0] = byte(a >> 24)
}
