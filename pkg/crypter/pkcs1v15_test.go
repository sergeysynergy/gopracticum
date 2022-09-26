// Package crypter Пакет реализует шифрование на основе ассиметричной пары ключей.
package crypter

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Этот файл реализует шифрование и дешифрование с использованием заполнения PKCS #1 v1.5.

func TestEncryptDecrypt(t *testing.T) {
	private, err := CreateKey(2 << 8)
	assert.NoError(t, err)

	privatePath := "/tmp/private.pem"
	pubPath := "/tmp/pub.pub"
	err = SavePemKeys(private, privatePath, pubPath)
	assert.NoError(t, err)

	privateFromFile, err := OpenPrivate(privatePath)
	assert.NoError(t, err)
	assert.Equal(t, private, privateFromFile)

	pubFromFile, err := OpenPublic(pubPath)
	assert.NoError(t, err)
	assert.Equal(t, &private.PublicKey, pubFromFile)

	//secretMessage := []byte("send reinforcements, we're going to advance, to big advance")
	secretMessage := []byte("send reinforcements, we're going to advance")
	cipherText, err := Encrypt(&private.PublicKey, secretMessage)
	assert.NoError(t, err)

	plainText, err := Decrypt(private, cipherText)
	assert.NoError(t, err)
	assert.Equal(t, secretMessage, plainText)

	fmt.Printf("Plain text: %s\n", plainText)

	//os.Remove(privatePath)
	//os.Remove(pubPath)
}
