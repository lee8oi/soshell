/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

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

func getStream(key []byte) cipher.Stream {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	var iv [aes.BlockSize]byte
	return cipher.NewOFB(block, iv[:])
}

func readFile(key []byte, path string) (b []byte, e error) {
	file, e := os.Open(path)
	if e == nil {
		defer file.Close()
		b, e = ioutil.ReadAll(&cipher.StreamReader{S: getStream(key), R: file})
	}
	return
}

func writeFile(key []byte, path string, data []byte) (e error) {
	file, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if e == nil {
		defer file.Close()
		writer := &cipher.StreamWriter{S: getStream(key), W: file}
		_, e = writer.Write(data)
	}
	return
}

func saveObject(obj interface{}, path, password string) (e error) {
	b, e := json.Marshal(obj)
	if e == nil {
		key := []byte(fmt.Sprintf("%x", sha256.Sum256([]byte(password)))[:32])
		e = writeFile(key, path, b)
	}
	return
}

func loadObject(obj interface{}, path, password string) (e error) {
	key := []byte(fmt.Sprintf("%x", sha256.Sum256([]byte(password)))[:32])
	b, e := readFile(key, path)
	if e == nil {
		e = json.Unmarshal(b, &obj)
	}
	return
}
