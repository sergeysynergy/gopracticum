// Package crypter Пакет реализует шифрование на основе ассиметричной пары ключей.
package crypter

import "testing"

// Этот файл реализует шифрование и дешифрование с использованием заполнения PKCS #1 v1.5.

func TestEncryptDecrypt(t *testing.T) {
	CreatePair(2 << 11)

	//pubKey := openPublic()
	//secretMessage := []byte("send reinforcements, we're going to advance")
	//cipherText, err := encrypt(pubKey, secretMessage)
	//if err != nil {
	//	log.Fatalln("Failed to encrypt", err)
	//}
	//fmt.Printf("Ciphertext: %x\n", cipherText)

	//privKey := openPrivate()
	//plainText, err := decrypt(privKey, cipherText)
	//if err != nil {
	//	log.Fatalln("Failed to decrypt", err)
	//}
	//fmt.Printf("Plain text: %s\n", plainText)
	//fmt.Printf("Plain text: %x\n", string(plainText))
}
