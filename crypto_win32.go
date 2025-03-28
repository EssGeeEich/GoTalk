package main

import "github.com/billgraziano/dpapi"

func crypto_decrypt(encryptedInput string) (string, error) {
	return dpapi.Decrypt(encryptedInput)
}

func crypto_encrypt(rawInput string) (string, error) {
	return dpapi.Encrypt(rawInput)
}
