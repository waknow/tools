package tripledes

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"errors"
)

type Mode uint

//des加密模式
const (
	ECB Mode = iota + 1
	BCB
)

type Padding uint

const (
	PCKS5 Padding = iota + 1
	Zero
)

//des加密，mode为模式
func DesEncrypt(origData, key []byte, mode Mode, padding Padding) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if mode == ECB {
		return encryptECB(origData, key, block, padding)
	}
	return encryptBCB(origData, key, block, padding)
}

//3des加密，mode为模式
func TripleDesEncrypt(origData, key []byte, mode Mode, padding Padding) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}

	if mode == ECB {
		return encryptECB(origData, key, block, padding)
	}
	return encryptBCB(origData, key, block, padding)
}

func encryptBCB(origData, key []byte, block cipher.Block, padding Padding) ([]byte, error) {
	switch padding {
	case PCKS5:
		origData = PKCS5Padding(origData, block.BlockSize())
	case Zero:
		origData = ZeroPadding(origData, block.BlockSize())
	default:
		return nil, errors.New("padding type error!")
	}

	crypted := make([]byte, len(origData))
	blockMode := cipher.NewCBCEncrypter(block, key)
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func encryptECB(origData, key []byte, block cipher.Block, padding Padding) ([]byte, error) {
	switch padding {
	case PCKS5:
		origData = PKCS5Padding(origData, block.BlockSize())
	case Zero:
		origData = ZeroPadding(origData, block.BlockSize())
	default:
		return nil, errors.New("padding type error!")
	}

	bs := block.BlockSize()
	if len(origData)%bs != 0 {
		return nil, errors.New("Need a multiple of the blocksize.")
	}
	crypted := make([]byte, len(origData))
	dst := crypted
	for len(origData) > 0 {
		block.Encrypt(dst, origData)
		origData = origData[bs:]
		dst = dst[bs:]
	}
	return crypted, nil
}

//des解密，mode为模式
func DesDecrypt(crypted, key []byte, mode Mode, padding Padding) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if mode == ECB {
		return decryptECB(crypted, key, block, padding)
	}
	return decryptBCB(crypted, key, block, padding)
}

//3des解密，mode为模式
func TripleDesDecrypt(crypted, key []byte, mode Mode, padding Padding) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}

	if mode == ECB {
		return decryptECB(crypted, key, block, padding)
	}
	return decryptBCB(crypted, key, block, padding)
}

func decryptECB(crypted, key []byte, block cipher.Block, padding Padding) ([]byte, error) {
	origin := make([]byte, len(crypted))
	dst := origin
	bs := block.BlockSize()
	if len(crypted)%bs != 0 {
		return nil, errors.New("crypto/cipher:input not full block!")
	}

	for len(crypted) > 0 {
		block.Decrypt(dst, crypted[:bs])
		crypted = crypted[bs:]
		dst = dst[bs:]
	}

	switch padding {
	case PCKS5:
		origin = PKCS5UnPadding(origin)
	case Zero:
		origin = ZeroUnPadding(origin)
	default:
		return nil, errors.New("padding type error!")
	}
	return origin, nil
}

func decryptBCB(crypted, key []byte, block cipher.Block, padding Padding) ([]byte, error) {
	blockMode := cipher.NewCBCDecrypter(block, key)
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)

	switch padding {
	case PCKS5:
		origData = PKCS5UnPadding(origData)
	case Zero:
		origData = ZeroUnPadding(origData)
	default:
		return nil, errors.New("padding type error!")
	}
	return origData, nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

func ZeroUnPadding(origData []byte) []byte {
	return bytes.TrimFunc(origData, func(r rune) bool {
		return r == rune(0)
	})
}
