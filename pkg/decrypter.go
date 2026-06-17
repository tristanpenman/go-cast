package internal

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
)

type Decrypter struct {
	aesIvMask []byte
	block     cipher.Block
	stream    cipher.Stream
}

func (decrypter *Decrypter) Decrypt(payload []byte, output []byte) int {
	decrypter.stream.XORKeyStream(output, payload)
	return len(output)
}

func (decrypter *Decrypter) Reset(frameNumber int) {
	left := new(bytes.Buffer)
	err := binary.Write(left, binary.BigEndian, uint32(frameNumber))
	if err != nil {
		return
	}

	iv := new(bytes.Buffer)
	iv.Write(decrypter.aesIvMask)

	for i := range left.Bytes() {
		iv.Bytes()[i+8] = left.Bytes()[i] ^ iv.Bytes()[8+i]
	}

	decrypter.stream = cipher.NewCTR(decrypter.block, iv.Bytes())
}

func NewDecrypter(aesKey []byte, aesIvMask []byte) *Decrypter {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil
	}

	stream := cipher.NewCTR(block, aesIvMask)

	return &Decrypter{
		aesIvMask: aesIvMask,
		block:     block,
		stream:    stream,
	}
}
