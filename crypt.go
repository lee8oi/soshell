/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

/*
The crypt system converts runtime data objects to/from json data stored in
password encrypted files. Encryption is done through cipher stream's which decrypt
or encrypt data during file reading or writing.
*/

//
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	//"log"
	"os"
)

// saveObject converts a data object to json and writes it to a file encrypted
// with specified password.
func saveObject(obj interface{}, path, password string) (e error) {
	b, e := json.Marshal(obj)
	if e == nil {
		key := []byte(fmt.Sprintf("%x", sha256.Sum256([]byte(password)))[:32])
		e = writeFile(key, path, b)
	}
	return
}

// loadObject reads json data from password encrypted file into a data object.
func loadObject(obj interface{}, path, password string) (e error) {
	key := []byte(fmt.Sprintf("%x", sha256.Sum256([]byte(password)))[:32])
	b, e := readFile(key, path)
	if e == nil {
		e = json.Unmarshal(b, &obj)
	}
	return
}

// getStream gets a new cipher stream for key.
func getStream(key []byte) cipher.Stream {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	var iv [aes.BlockSize]byte
	return cipher.NewOFB(block, iv[:])
}

// readFile decrypts a file as it reads the specified file's data.
func readFile(key []byte, path string) (b []byte, e error) {
	file, e := os.Open(path)
	if e == nil {
		defer file.Close()
		b, e = ioutil.ReadAll(&cipher.StreamReader{S: getStream(key), R: file})
	}
	return
}

// writeFile encrypts data as it writes to the specified file.
func writeFile(key []byte, path string, data []byte) (e error) {
	file, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if e == nil {
		defer file.Close()
		writer := &cipher.StreamWriter{S: getStream(key), W: file}
		_, e = writer.Write(data)
	}
	return
}
