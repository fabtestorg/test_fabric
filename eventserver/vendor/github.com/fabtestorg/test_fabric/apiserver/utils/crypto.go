package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"math/rand"
	"os"
	"time"

	crypto "github.com/hyperledger/fabric/core/crypto"
	"github.com/hyperledger/fabric/core/crypto/primitives"
	"github.com/hyperledger/fabric/core/crypto/primitives/ecies"
)

func init() {
	if err := crypto.Init(); err != nil {
		panic("crypto init err : " + err.Error())
	}
}

func GetPublicKey(pem []byte) (interface{}, error) {
	publicKey, err := primitives.PEMtoPublicKey(pem, nil)
	return publicKey, err
}

func GetPrivateKey(pem []byte) (interface{}, error) {
	privatekey, err := primitives.PEMtoPrivateKey(pem, nil)
	return privatekey, err
}

func IsFileExist(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ReadFile(path string) ([]byte, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("ReadFile open file [%s] err : %s\n", path, err.Error())
		return nil, err
	}
	info, err := os.Stat(path)
	fileSize := info.Size()
	buff := make([]byte, fileSize)
	_, err = f.Read(buff)
	if err != nil {
		fmt.Printf("ReadFile read file [%s] err : %s\n", path, err.Error())
		return nil, err
	}

	f.Close()

	return buff, nil
}

func GenerateKey(bitsize int) []byte {
	key := ""

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < bitsize/2; i++ {
		key += fmt.Sprintf("%x", r.Int63())
	}

	return []byte(key)
}

func AesEncrypt(origData, key []byte) ([]byte, error) {
	key = primitives.HMACAESTruncated(key, []byte{1})
	var iv = []byte(key)[:aes.BlockSize]
	encrypted := make([]byte, len(origData))
	aesBlockEncrypter, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncrypter, iv)
	aesEncrypter.XORKeyStream(encrypted, origData)
	return encrypted, nil
}

func AesDecrypt(crypted, key []byte) (decrypted []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	key = primitives.HMACAESTruncated(key, []byte{1})
	var iv = []byte(key)[:aes.BlockSize]
	decrypted = make([]byte, len(crypted))
	var aesBlockDecrypter cipher.Block
	aesBlockDecrypter, err = aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, iv)
	aesDecrypter.XORKeyStream(decrypted, crypted)
	return decrypted, nil
}

func EciesEncrypt(originData, key []byte) ([]byte, error) {
	cert, _, err := primitives.PEMtoCertificateAndDER(key)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	spi := ecies.NewSPI()

	tmpPubKey, err := spi.NewPublicKey(nil, cert.PublicKey)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	publicKey, err := spi.NewAsymmetricCipherFromPublicKey(tmpPubKey)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	cryptoData, err := publicKey.Process([]byte(originData))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return cryptoData, nil
}

func EciesDecrypt(cryptoData, key []byte) ([]byte, error) {
	tmpPrivateKey, err := primitives.PEMtoPrivateKey(key, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	spi := ecies.NewSPI()

	tmpPriKey, err := spi.NewPrivateKey(nil, tmpPrivateKey)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	privateKey, err := spi.NewAsymmetricCipherFromPrivateKey(tmpPriKey)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	srcData, err := privateKey.Process(cryptoData)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return srcData, nil
}
