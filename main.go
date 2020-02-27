package main

import (
	"github.com/DmitriiTrifonov/magmafier-bot/magmafier"
	tb "gopkg.in/tucnak/telebot.v2"
	"image"
	"image/color"
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

	message := "This is the Mamgafier bot.\n" +
		"It uses block cipher \"Mamga\" from GOST 34.12-2018.\n" +
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
		log.Println(caption)
		mgm.SetKey(magmafier.MakeKeyFromString(caption))
		photoUrl := m.Photo.File.FileID
		url, err := b.FileURLByID(photoUrl)
		log.Println(url)
		resp, err := http.Get(url)
		img, _, err := image.Decode(resp.Body)
		x := img.Bounds().Dx()
		y := img.Bounds().Dy()
		err = resp.Body.Close()
		mod := image.NewRGBA(image.Rect(0, 0, x, y))
		for i := 0; i < x; i++ {
			for j := 0; j < y; j++ {
				c := img.At(i, j)
				r, g, b, a := c.RGBA()

				var rArr [4]byte
				convertToArray(r, &rArr)
				log.Println("rArr:", rArr)
				var gArr [4]byte
				convertToArray(g, &gArr)
				log.Println("gArr:", gArr)
				rgCipher := make([]byte, 0)
				rgCipher = append(rgCipher, rArr[:]...)
				rgCipher = append(rgCipher, gArr[:]...)

				log.Println("rgCipher:", rgCipher)

				var bArr [4]byte
				convertToArray(b, &bArr)
				var aArr [4]byte
				convertToArray(a, &aArr)
				baCipher := make([]byte, 0)
				baCipher = append(baCipher, bArr[:]...)
				baCipher = append(baCipher, aArr[:]...)

				log.Println("baCipher:", baCipher)

				rgCipher = mgm.Encrypt(rgCipher)
				baCipher = mgm.Encrypt(baCipher)

				log.Println("rgCipher:", rgCipher)
				log.Println("baCipher:", baCipher)

				var rgArr, baArr [8]byte
				copy(rgArr[:], rgCipher)
				copy(baArr[:], baCipher)
				newR, newG := convertToUInt32(&rgArr)
				newB, newA := convertToUInt32(&baArr)
				log.Println(r, g, b, a)
				log.Println(newR, newG, newB, newA)
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
