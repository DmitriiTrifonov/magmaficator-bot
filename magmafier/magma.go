package magmafier

type Magma struct {
	key     [0x20]byte
	subKeys [0x20][4]byte
}

func (m *Magma) ResetSubKeys() {
	for i := 0; i < 32; i++ {
		for j := 0; j < 8; j++ {
			m.subKeys[i][j] = 0
		}
	}
}

var pi = [8][16]byte{
	{1, 7, 14, 13, 0, 5, 8, 3, 4, 15, 10, 6, 9, 12, 11, 2},
	{8, 14, 2, 5, 6, 9, 1, 12, 15, 4, 11, 0, 13, 10, 3, 7},
	{5, 13, 15, 6, 9, 2, 12, 10, 11, 7, 8, 1, 4, 3, 14, 0},
	{7, 15, 5, 10, 8, 1, 6, 13, 0, 9, 3, 14, 11, 4, 2, 12},
	{12, 8, 2, 1, 13, 4, 15, 6, 7, 0, 10, 5, 3, 14, 9, 11},
	{11, 3, 5, 8, 2, 15, 10, 13, 14, 1, 7, 4, 12, 9, 6, 0},
	{6, 8, 2, 3, 9, 10, 5, 12, 1, 14, 4, 7, 11, 13, 0, 15},
	{12, 4, 6, 2, 10, 5, 11, 9, 14, 8, 13, 7, 0, 3, 15, 1},
}

func xor(a *[4]byte, b *[4]byte, o *[4]byte) {
	for i := 0; i < 4; i++ {
		o[i] = a[i] ^ b[i]
	}
}

func convertToUInt32(a *[4]byte) uint32 {
	var r uint32
	for i := 0; i < 3; i++ {
		r |= uint32(a[i])
		r <<= 8
	}
	r |= uint32(a[3])
	return r
}

func convertToArray(a uint32, arr *[4]byte) {
	arr[3] = byte(a)
	arr[2] = byte(a >> 8)
	arr[1] = byte(a >> 16)
	arr[0] = byte(a >> 24)
}

func x32(a *[4]byte, b *[4]byte, o *[4]byte) {
	var internal int
	for i := 3; i >= 0; i-- {
		internal = int(a[i]) + int(b[i]) + (internal >> 8)
		o[i] = byte(internal & 0xFF)
	}
}

// Splits bytes in array by two 4 bits numbers and changes value from pi table
func t(input *[4]byte, out *[4]byte) {
	var fbp, sbp byte
	for i := 0; i < 4; i++ {
		fbp = (input[i] & 0xF0) >> 4
		sbp = input[i] & 0x0F
		fbp = pi[i*2][fbp]
		sbp = pi[i*2+1][sbp]
		out[i] = (fbp << 4) | sbp
	}
}

func (m *Magma) SetSubKeys() {
	copy(m.subKeys[0][:], m.key[:4])
	copy(m.subKeys[1][:], m.key[4:8])
	copy(m.subKeys[2][:], m.key[8:12])
	copy(m.subKeys[3][:], m.key[12:16])
	copy(m.subKeys[4][:], m.key[16:20])
	copy(m.subKeys[5][:], m.key[20:24])
	copy(m.subKeys[6][:], m.key[24:28])
	copy(m.subKeys[7][:], m.key[28:])
	copy(m.subKeys[8][:], m.key[:4])
	copy(m.subKeys[9][:], m.key[4:8])
	copy(m.subKeys[10][:], m.key[8:12])
	copy(m.subKeys[11][:], m.key[12:16])
	copy(m.subKeys[12][:], m.key[16:20])
	copy(m.subKeys[13][:], m.key[20:24])
	copy(m.subKeys[14][:], m.key[24:28])
	copy(m.subKeys[15][:], m.key[28:])
	copy(m.subKeys[16][:], m.key[:4])
	copy(m.subKeys[17][:], m.key[4:8])
	copy(m.subKeys[18][:], m.key[8:12])
	copy(m.subKeys[19][:], m.key[12:16])
	copy(m.subKeys[20][:], m.key[16:20])
	copy(m.subKeys[21][:], m.key[20:24])
	copy(m.subKeys[22][:], m.key[24:28])
	copy(m.subKeys[23][:], m.key[28:])
	copy(m.subKeys[24][:], m.key[28:])
	copy(m.subKeys[25][:], m.key[24:28])
	copy(m.subKeys[26][:], m.key[20:24])
	copy(m.subKeys[27][:], m.key[16:20])
	copy(m.subKeys[28][:], m.key[12:16])
	copy(m.subKeys[29][:], m.key[8:12])
	copy(m.subKeys[30][:], m.key[4:8])
	copy(m.subKeys[31][:], m.key[:4])
}

func gSwap(block *[4]byte, key *[4]byte, out *[4]byte) {
	var out32 [4]byte
	x32(block, key, &out32)
	var outT [4]byte
	t(&out32, &outT)
	n := convertToUInt32(&outT)
	n = (n << 11) | (n >> 21)
	convertToArray(n, out)
}

func gIter(block *[8]byte, key *[4]byte, out *[8]byte) {
	var rh [4]byte
	var lh [4]byte
	var G [4]byte

	for i := 0; i < 4; i++ {
		rh[i] = block[4+i]
		lh[i] = block[i]
	}

	gSwap(key, &rh, &G)
	xor(&G, &lh, &G)

	for i := 0; i < 4; i++ {
		lh[i] = rh[i]
		rh[i] = G[i]
	}

	for i := 0; i < 4; i++ {
		out[i] = lh[i]
		out[4+i] = rh[i]
	}
}

func gFinal(block *[8]byte, key *[4]byte, out *[8]byte) {
	var rh [4]byte
	var lh [4]byte
	var G [4]byte

	for i := 0; i < 4; i++ {
		rh[i] = block[4+i]
		lh[i] = block[i]
	}

	gSwap(key, &rh, &G)
	xor(&G, &lh, &G)

	for i := 0; i < 4; i++ {
		lh[i] = G[i]
	}

	for i := 0; i < 4; i++ {
		out[i] = lh[i]
		out[4+i] = rh[i]
	}
}

func (m *Magma) Encrypt(data []byte) []byte {
	var arr [8]byte
	var out [8]byte
	copy(arr[:], data)
	block := arr
	gIter(&block, &m.subKeys[0], &out)

	for i := 1; i < 31; i++ {
		gIter(&out, &m.subKeys[i], &out)
	}

	gFinal(&out, &m.subKeys[31], &out)

	return out[:]
}

func (m *Magma) Decrypt(data []byte) []byte {
	var arr [8]byte
	var out [8]byte
	copy(arr[:], data)
	block := arr

	gIter(&block, &m.subKeys[31], &out)

	for i := 30; i > 0; i-- {
		gIter(&out, &m.subKeys[i], &out)
	}

	gFinal(&out, &m.subKeys[0], &out)

	return out[:]
}

func (m *Magma) SetKey(data []byte) {
	var arr [0x20]byte
	copy(arr[:], data[:0x20])
	m.key = arr
}
