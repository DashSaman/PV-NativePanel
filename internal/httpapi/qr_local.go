package httpapi

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/png"
)

const (
	qrVersion              = 6
	qrSize                 = 41
	qrDataCodewords        = 136
	qrDataBytesPerBlock    = 68
	qrECCCodewordsPerBlock = 18
	qrMaxByteLength        = 134
)

func qrAppendBits(bits *[]bool, value, length int) {
	for i := length - 1; i >= 0; i-- {
		*bits = append(*bits, ((value>>i)&1) != 0)
	}
}

func qrMultiplyGF(x, y int) int {
	z := 0
	for i := 7; i >= 0; i-- {
		z = (z << 1) ^ (((z >> 7) & 1) * 0x11d)
		z ^= ((y >> i) & 1) * x
	}
	return z
}

func qrComputeDivisor(degree int) []byte {
	result := make([]byte, degree)
	result[degree-1] = 1
	root := 1
	for i := 0; i < degree; i++ {
		for j := 0; j < degree; j++ {
			v := qrMultiplyGF(int(result[j]), root)
			if j+1 < degree {
				v ^= int(result[j+1])
			}
			result[j] = byte(v)
		}
		root = qrMultiplyGF(root, 0x02)
	}
	return result
}

func qrComputeRemainder(data, divisor []byte) []byte {
	result := make([]byte, len(divisor))
	for _, value := range data {
		factor := int(value ^ result[0])
		copy(result, result[1:])
		result[len(result)-1] = 0
		for i := range divisor {
			result[i] ^= byte(qrMultiplyGF(int(divisor[i]), factor))
		}
	}
	return result
}

func qrBuildDataCodewords(text string) ([]byte, error) {
	payload := []byte(text)
	if len(payload) > qrMaxByteLength {
		return nil, errors.New("qr payload is too long")
	}
	bits := make([]bool, 0, qrDataCodewords*8)
	qrAppendBits(&bits, 0b0100, 4)
	qrAppendBits(&bits, len(payload), 8)
	for _, value := range payload {
		qrAppendBits(&bits, int(value), 8)
	}
	capacity := qrDataCodewords * 8
	terminator := 4
	if capacity-len(bits) < terminator {
		terminator = capacity - len(bits)
	}
	for i := 0; i < terminator; i++ {
		bits = append(bits, false)
	}
	for len(bits)%8 != 0 {
		bits = append(bits, false)
	}
	data := make([]byte, 0, qrDataCodewords)
	for i := 0; i < len(bits); i += 8 {
		value := 0
		for j := 0; j < 8; j++ {
			value <<= 1
			if bits[i+j] {
				value |= 1
			}
		}
		data = append(data, byte(value))
	}
	for pad := 0; len(data) < qrDataCodewords; pad++ {
		if pad%2 == 0 {
			data = append(data, 0xec)
		} else {
			data = append(data, 0x11)
		}
	}
	return data, nil
}

func qrBuildCodewords(text string) ([]byte, error) {
	data, err := qrBuildDataCodewords(text)
	if err != nil {
		return nil, err
	}
	divisor := qrComputeDivisor(qrECCCodewordsPerBlock)
	blocks := [][]byte{data[:qrDataBytesPerBlock], data[qrDataBytesPerBlock:]}
	ecc := [][]byte{qrComputeRemainder(blocks[0], divisor), qrComputeRemainder(blocks[1], divisor)}
	out := make([]byte, 0, qrDataCodewords+2*qrECCCodewordsPerBlock)
	for i := 0; i < qrDataBytesPerBlock; i++ {
		out = append(out, blocks[0][i], blocks[1][i])
	}
	for i := 0; i < qrECCCodewordsPerBlock; i++ {
		out = append(out, ecc[0][i], ecc[1][i])
	}
	return out, nil
}

func qrFormatBits(mask int) int {
	data := (1 << 3) | mask
	remainder := data
	for i := 0; i < 10; i++ {
		remainder = (remainder << 1) ^ (((remainder >> 9) & 1) * 0x537)
	}
	return ((data << 10) | remainder) ^ 0x5412
}

func encodeLocalQR(text string) ([][]bool, error) {
	if qrVersion != 6 {
		return nil, errors.New("unsupported qr version")
	}
	modules := make([][]bool, qrSize)
	functions := make([][]bool, qrSize)
	for i := 0; i < qrSize; i++ {
		modules[i] = make([]bool, qrSize)
		functions[i] = make([]bool, qrSize)
	}
	setFunction := func(x, y int, dark bool) {
		if x < 0 || y < 0 || x >= qrSize || y >= qrSize {
			return
		}
		modules[y][x] = dark
		functions[y][x] = true
	}
	for i := 8; i < qrSize-8; i++ {
		setFunction(6, i, i%2 == 0)
		setFunction(i, 6, i%2 == 0)
	}
	centers := [][2]int{{3, 3}, {qrSize - 4, 3}, {3, qrSize - 4}}
	for _, center := range centers {
		for dy := -4; dy <= 4; dy++ {
			for dx := -4; dx <= 4; dx++ {
				d := absInt(dx)
				if absInt(dy) > d {
					d = absInt(dy)
				}
				setFunction(center[0]+dx, center[1]+dy, d != 2 && d != 4)
			}
		}
	}
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			d := absInt(dx)
			if absInt(dy) > d {
				d = absInt(dy)
			}
			setFunction(34+dx, 34+dy, d != 1)
		}
	}
	format := qrFormatBits(0)
	bit := func(index int) bool { return ((format >> index) & 1) != 0 }
	for i := 0; i <= 5; i++ {
		setFunction(8, i, bit(i))
	}
	setFunction(8, 7, bit(6))
	setFunction(8, 8, bit(7))
	setFunction(7, 8, bit(8))
	for i := 9; i < 15; i++ {
		setFunction(14-i, 8, bit(i))
	}
	for i := 0; i < 8; i++ {
		setFunction(qrSize-1-i, 8, bit(i))
	}
	for i := 8; i < 15; i++ {
		setFunction(8, qrSize-15+i, bit(i))
	}
	setFunction(8, qrSize-8, true)

	codewords, err := qrBuildCodewords(text)
	if err != nil {
		return nil, err
	}
	bitIndex := 0
	for right := qrSize - 1; right >= 1; right -= 2 {
		if right == 6 {
			right = 5
		}
		upward := ((right + 1) & 2) == 0
		for vertical := 0; vertical < qrSize; vertical++ {
			y := vertical
			if upward {
				y = qrSize - 1 - vertical
			}
			for j := 0; j < 2; j++ {
				x := right - j
				if functions[y][x] || bitIndex >= len(codewords)*8 {
					continue
				}
				modules[y][x] = ((codewords[bitIndex>>3] >> (7 - (bitIndex & 7))) & 1) != 0
				bitIndex++
			}
		}
	}
	if bitIndex != len(codewords)*8 {
		return nil, errors.New("qr data placement failed")
	}
	for y := 0; y < qrSize; y++ {
		for x := 0; x < qrSize; x++ {
			if !functions[y][x] && (x+y)%2 == 0 {
				modules[y][x] = !modules[y][x]
			}
		}
	}
	return modules, nil
}

func localQRDataURI(text string) (string, error) {
	matrix, err := encodeLocalQR(text)
	if err != nil {
		return "", err
	}
	const quiet = 4
	const scale = 6
	side := (qrSize + quiet*2) * scale
	img := image.NewRGBA(image.Rect(0, 0, side, side))
	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 255}
	for y := 0; y < side; y++ {
		for x := 0; x < side; x++ {
			img.SetRGBA(x, y, white)
		}
	}
	for y, row := range matrix {
		for x, dark := range row {
			if !dark {
				continue
			}
			for py := 0; py < scale; py++ {
				for px := 0; px < scale; px++ {
					img.SetRGBA((x+quiet)*scale+px, (y+quiet)*scale+py, black)
				}
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
