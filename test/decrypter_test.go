package test

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"testing"

	// internal
	. "github.com/tristanpenman/go-cast/internal"
)

func TestNewDecrypterInvalidKeyLength(t *testing.T) {
	if decrypter := NewDecrypter([]byte{0x00, 0x01, 0x02}, make([]byte, aes.BlockSize)); decrypter != nil {
		t.Fatalf("expected nil decrypter for invalid key length, got %#v", decrypter)
	}
}

func TestDecryptWithoutReset(t *testing.T) {
	key := []byte("example key 1234")
	ivMask := []byte{
		0x01, 0x23, 0x45, 0x67,
		0x89, 0xab, 0xcd, 0xef,
		0x10, 0x32, 0x54, 0x76,
		0x98, 0xba, 0xdc, 0xfe,
	}

	plaintext := []byte("hello world!!!!")
	ciphertext := make([]byte, len(plaintext))

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("unable to create AES block: %v", err)
	}

	reference := cipher.NewCTR(block, ivMask)
	reference.XORKeyStream(ciphertext, plaintext)

	decrypter := NewDecrypter(key, ivMask)
	if decrypter == nil {
		t.Fatal("expected valid decrypter")
	}

	output := make([]byte, len(ciphertext))
	if n := decrypter.Decrypt(ciphertext, output); n != len(output) {
		t.Fatalf("unexpected output length: got %d want %d", n, len(output))
	}

	if !bytes.Equal(output, plaintext) {
		t.Fatalf("unexpected plaintext: got %x want %x", output, plaintext)
	}
}

func TestDecryptAfterReset(t *testing.T) {
	key := []byte("example key 1234")
	ivMask := []byte{
		0x00, 0x11, 0x22, 0x33,
		0x44, 0x55, 0x66, 0x77,
		0x88, 0x99, 0xaa, 0xbb,
		0xcc, 0xdd, 0xee, 0xff,
	}

	frameNumber := 42

	iv := append([]byte{}, ivMask...)
	frameBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(frameBytes, uint32(frameNumber))
	for i := range frameBytes {
		iv[8+i] = frameBytes[i] ^ iv[8+i]
	}

	plaintext := []byte("frame-specific data")
	ciphertext := make([]byte, len(plaintext))

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("unable to create AES block: %v", err)
	}

	reference := cipher.NewCTR(block, iv)
	reference.XORKeyStream(ciphertext, plaintext)

	decrypter := NewDecrypter(key, ivMask)
	if decrypter == nil {
		t.Fatal("expected valid decrypter")
	}

	decrypter.Reset(frameNumber)

	output := make([]byte, len(ciphertext))
	if n := decrypter.Decrypt(ciphertext, output); n != len(output) {
		t.Fatalf("unexpected output length: got %d want %d", n, len(output))
	}

	if !bytes.Equal(output, plaintext) {
		t.Fatalf("unexpected plaintext: got %x want %x", output, plaintext)
	}
}
