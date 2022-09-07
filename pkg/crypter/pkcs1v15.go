// Package crypter Пакет реализует шифрование на основе ассиметричной пары ключей.
package crypter

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// Этот файл реализует шифрование и дешифрование с использованием заполнения PKCS #1 v1.5.

// Encrypt Шифруем набор байт посредством публичного ключа.
func Encrypt(pubKey *rsa.PublicKey, msg []byte) ([]byte, error) {
	// crypto/rand.Reader источник криптостойкой случайной последовательности
	rnd := rand.Reader

	buf := bytes.Buffer{}
	w := bufio.NewWriterSize(&buf, 32)
	// Сообщение должно быть не длиннее, чем длина публичного ключа минус 11 байт.
	cipherText, err := rsa.EncryptPKCS1v15(rnd, pubKey, msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from encryption: %s\n", err)
		return nil, err
	}
	w.Write(cipherText)

	// Так как используется источник случайной последовательности байт, результат всё время будет разный.
	return buf.Bytes(), nil
}

// Decrypt Дешифруем набор байт посредством приватного ключа.
func Decrypt(privKey *rsa.PrivateKey, cipherText []byte) ([]byte, error) {
	// Если rand != nil, он используется для защиты от `side-channel` атак.
	plainText, err := rsa.DecryptPKCS1v15(nil, privKey, cipherText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from encryption: %s\n", err)
		return nil, err
	}

	// Так как используется источник случайной последовательности байт, результат всё время будет разный.
	return plainText, nil
}

// CreateKey создаёт приватный RSA-ключ указанного размера, который содержит в себе публичную пару.
func CreateKey(keySize int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, keySize)
}

// SavePemKeys Записывает в файлы пару ассиметричных RSA-ключей в формате pem.
func SavePemKeys(privateKey *rsa.PrivateKey, privateKeyPath, publicKeyPath string) (err error) {
	// Сохраним в файл приватный ключ.
	pemPrivateFile, err := os.Create(privateKeyPath)
	defer pemPrivateFile.Close()
	if err != nil {
		return err
	}

	var pemPrivateBlock = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	err = pem.Encode(pemPrivateFile, pemPrivateBlock)
	if err != nil {
		return err
	}

	// Сохраним в файл публичный ключ.
	pemPublicFile, err := os.Create(publicKeyPath)
	defer pemPublicFile.Close()
	if err != nil {
		return err
	}

	var pemPublicBlock = &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	}

	err = pem.Encode(pemPublicFile, pemPublicBlock)
	if err != nil {
		return err
	}

	return nil
}

// OpenPrivate получает приватный ключ шифрования из файла.
func OpenPrivate(fileName string) (*rsa.PrivateKey, error) {
	privateKeyFile, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open private key file: %w", err)
	}

	pemfileinfo, err := privateKeyFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to open private key file: %w", err)
	}

	var size = pemfileinfo.Size()
	pembytes := make([]byte, size)
	buffer := bufio.NewReader(privateKeyFile)
	_, err = buffer.Read(pembytes)
	data, _ := pem.Decode([]byte(pembytes))
	privateKeyFile.Close()

	privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key file: %w", err)
	}

	return privateKeyImported, nil
}

// OpenPublic получает публичный ключ шифрования из файла.
func OpenPublic(fileName string) (*rsa.PublicKey, error) {
	publicKeyFile, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open public key file: %w", err)
	}
	defer publicKeyFile.Close()

	pemfileinfo, err := publicKeyFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to open public key file: %w", err)
	}

	var size = pemfileinfo.Size()
	pembytes := make([]byte, size)
	buffer := bufio.NewReader(publicKeyFile)
	_, err = buffer.Read(pembytes)
	data, _ := pem.Decode([]byte(pembytes))

	publicKeyImported, err := x509.ParsePKCS1PublicKey(data.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key file: %w", err)
	}

	return publicKeyImported, nil
}
