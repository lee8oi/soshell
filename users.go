/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

//
package main

import (
	//"errors"
	"log"
	//"os"
	"regexp"
)

type user struct {
	Email, Name string
}

// isEmail makes she that email is properly formated as an email address.
func isEmail(email string) bool {
	reg := regexp.MustCompile("^([\\w\\.\\-_]+)?\\w+@[\\w-_]+(\\.\\w+){1,}$")
	return reg.MatchString(email)
}

// isName checks if name only contains word characters.
func isName(name string) bool {
	nameReg := regexp.MustCompile("[\\W]+")
	return !nameReg.MatchString(name)
}

// load is used to load a users info from json stored in an encrypted file.
func (u *user) load(name, pass string) error {
	path := *users + SEP + indexPath([]byte(name))
	return loadObject(u, path+SEP+"user", pass)
}

// save will save a users info as json in an encrypted file.
func (u *user) save(name, pass string) error {
	path := *users + SEP + indexPath([]byte(name))
	if !pathExists(path) {
		err := makePath(path, 0700)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return saveObject(u, path+SEP+"user", pass)
}
