package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

func padding(src []byte, blocksize int) []byte {
	padnum := blocksize - len(src)%blocksize
	pad := bytes.Repeat([]byte{byte(padnum)}, padnum)
	return append(src, pad...)
}

func unpadding(src []byte) []byte {
	n := len(src)
	unpadnum := int(src[n-1])
	return src[:n-unpadnum]
}

func DncryptAES(src []byte, key [16]byte) []byte {
	block, _ := aes.NewCipher(key[:])
	src = padding(src, block.BlockSize())
	blockmode := cipher.NewCBCEncrypter(block, key[:])
	blockmode.CryptBlocks(src, src)
	return src
}

func DecryptAES(src []byte, key [16]byte) []byte {
	block, _ := aes.NewCipher(key[:])
	blockmode := cipher.NewCBCDecrypter(block, key[:])
	blockmode.CryptBlocks(src, src)
	src = unpadding(src)
	return src
}
