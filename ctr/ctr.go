package ctr

import (
	"github.com/DmitriiTrifonov/magmafier-bot/magmafier"
	"math"
)

var Vector = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

func incCtr(ctr []byte) {
	internal := uint32(0)
	one := make([]byte, len(ctr))
	one[len(ctr)-1] = 0x01
	for i := len(ctr) - 1; i >= 0; i-- {
		internal = uint32(ctr[i]) + uint32(one[i]) + (internal >> 8)
		ctr[i] = byte(internal & 0xFF)
	}
}

func xor(a []byte, b []byte) []byte {
	m := math.Min(float64(len(a)), float64(len(b)))
	s := make([]byte, int(m))
	for i := 0; i < int(m); i++ {
		s[i] = a[i] ^ b[i]
	}
	return s
}

func CTRCrypt(blk []byte, ctr []byte, mgm *magmafier.Magma) []byte {
	c := mgm.Encrypt(ctr)
	incCtr(ctr)
	r := xor(c, blk)
	return r
}
