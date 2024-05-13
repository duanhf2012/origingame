package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

func SpecialAesEncrypt(orig string, key string) (string, error) {
	aesRet, err := AesEncrypt(orig, key)

	if err != nil || len(aesRet) <= 2 {
		return aesRet, err
	}

	byteRet := []byte(aesRet)
	lenRet := len(byteRet)
	convertNum := (byteRet[0] % 2) == 0
	convertChar := (byteRet[1] % 2) != 0

	for i := 2; i < lenRet; i++ {
		if convertNum {
			switch byteRet[i] {
			case '0':
				byteRet[i] = '2'
			case '1':
				byteRet[i] = '7'
			case '2':
				byteRet[i] = '6'
			case '3':
				byteRet[i] = '5'
			case '4':
				byteRet[i] = '9'
			case '5':
				byteRet[i] = '4'
			case '6':
				byteRet[i] = '8'
			case '7':
				byteRet[i] = '3'
			case '8':
				byteRet[i] = '0'
			case '9':
				byteRet[i] = '1'
			}
		}

		if convertChar {
			if byteRet[i] >= 'a' && byteRet[i] <= 'z' {
				byteRet[i] -= 32
			} else if byteRet[i] >= 'A' && byteRet[i] <= 'Z' {
				byteRet[i] += 32
			}
		}
	}

	return string(byteRet), nil
}

func SpecialAesDecrypt(orig string, key string) (string, error) {
	if len(orig) <= 2 {
		return AesDecrypt(orig, key)
	}

	byteRet := []byte(orig)
	lenRet := len(byteRet)
	convertNum := (byteRet[0] % 2) == 0
	convertChar := (byteRet[1] % 2) != 0

	for i := 2; i < lenRet; i++ {
		if convertNum {
			switch byteRet[i] {
			case '0':
				byteRet[i] = '8'
			case '1':
				byteRet[i] = '9'
			case '2':
				byteRet[i] = '0'
			case '3':
				byteRet[i] = '7'
			case '4':
				byteRet[i] = '5'
			case '5':
				byteRet[i] = '3'
			case '6':
				byteRet[i] = '2'
			case '7':
				byteRet[i] = '1'
			case '8':
				byteRet[i] = '6'
			case '9':
				byteRet[i] = '4'
			}
		}

		if convertChar {
			if byteRet[i] >= 'a' && byteRet[i] <= 'z' {
				byteRet[i] -= 32
			} else if byteRet[i] >= 'A' && byteRet[i] <= 'Z' {
				byteRet[i] += 32
			}
		}
	}

	return AesDecrypt(string(byteRet), key)
}

func AesEncrypt(orig string, key string) (string, error) {
	// 转成字节数组
	origData := []byte(orig)
	k := []byte(key)

	// 分组秘钥
	block, err := aes.NewCipher(k)
	if err != nil {
		return "", err
	}

	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补全码
	origData = PKCS7Padding(origData, blockSize)
	if len(origData) == 0 {
		return "", fmt.Errorf("AesEncrypt fail orig %s", orig)
	}
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])
	if blockMode == nil {
		return "", fmt.Errorf("AesEncrypt fail orig %s", orig)
	}

	// 创建数组
	cryted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(cryted, origData)

	return base64.StdEncoding.EncodeToString(cryted), nil

}

func AesDecrypt(cryted string, key string) (string, error) {
	// 转成字节数组
	crytedByte, err := base64.StdEncoding.DecodeString(cryted)
	if err != nil {
		return "", err
	}

	k := []byte(key)
	// 分组秘钥
	block, err := aes.NewCipher(k)
	if err != nil {
		return "", err
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	if blockMode == nil {
		return "", fmt.Errorf("AesEncrypt fail orig %s", cryted)
	}
	// 创建数组
	orig := make([]byte, len(crytedByte))
	// 解密
	blockMode.CryptBlocks(orig, crytedByte)
	// 去补全码
	orig = PKCS7UnPadding(orig)
	if len(orig) == 0 {
		return "", fmt.Errorf("AesEncrypt fail orig %s", cryted)
	}

	return string(orig), nil
}

// 补码
func PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// 去码
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	if length <= 0 {
		return nil
	}
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

func ZeroUnPadding(origData []byte) []byte {
	return bytes.TrimFunc(origData,
		func(r rune) bool {
			return r == rune(0)
		})
}

func DesEncrypt(text string, key []byte) (string, error) {
	src := []byte(text)
	block, err := des.NewCipher(key)
	if err != nil {
		return "", err
	}
	bs := block.BlockSize()
	src = ZeroPadding(src, bs)
	if len(src)%bs != 0 {
		return "", errors.New("Need a multiple of the blocksize")
	}
	out := make([]byte, len(src))
	dst := out
	for len(src) > 0 {
		block.Encrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
	return hex.EncodeToString(out), nil
}

func DesDecrypt(decrypted string, key []byte) (string, error) {
	src, err := hex.DecodeString(decrypted)
	if err != nil {
		return "", err
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return "", err
	}
	out := make([]byte, len(src))
	dst := out
	bs := block.BlockSize()
	if len(src)%bs != 0 {
		return "", errors.New("crypto/cipher: input not full blocks")
	}
	for len(src) > 0 {
		block.Decrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
	out = ZeroUnPadding(out)
	return string(out), nil
}
