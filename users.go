/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

//
package main

import (
	"errors"
	//	"fmt"
	"log"
	//	"os"
	"encoding/json"
	"github.com/HouzuoGuo/tiedot/db"
	"regexp"
	"strings"
	// "github.com/HouzuoGuo/tiedot/dberr"
)

var (
	database *db.DB
	userDB   *db.Col
)

type user struct {
	Email, Name, DocID string
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

func loadUserDB() {
	var err error
	database, err = db.OpenDB(*work + SEP + "database")
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Create("users"); err == nil {
		userDB = database.Use("users")
		if err := userDB.Index([]string{"Name"}); err != nil {
			log.Println(err)
		}
		if err := userDB.Index([]string{"Pass"}); err != nil {
			log.Println(err)
		}
		if err := userDB.Index([]string{"Email"}); err != nil {
			log.Println(err)
		}
		log.Println("User database created.")
	} else {
		userDB = database.Use("users")
	}
	log.Println("Loaded user database.")
}

func closeUserDB() {
	if err := database.Close(); err != nil {
		log.Println(err)
	}
	log.Println("Closed user database.")
}

func userExists(name string) bool {
	rb, err := queryUser(name)
	if err != nil {
		log.Println(err)
	}
	log.Printf("user check: %v", rb)
	if rb["Name"] == strings.ToLower(name) {
		return true
	}
	return false
}

func queryUser(name string) (rb map[string]interface{}, err error) {
	var query interface{}
	result := make(map[int]struct{})
	json.Unmarshal([]byte(`[{"eq": "`+strings.ToLower(name)+`", "in": ["Name"]}]`), &query)
	if err := db.EvalQuery(query, userDB, &result); err != nil {
		log.Println(err)
		return nil, err
	}
	for id := range result {
		rb, err = userDB.Read(id)
		if err != nil {
			return nil, err
		}
		break //only need one result
	}
	log.Printf("%v", rb)
	return rb, nil
}

// login checks the users password and loads their info from the users database.
func (u *user) login(name, pass string) error {
	if rb, err := queryUser(name); err != nil {
		log.Println(err)
		return err
	} else {
		if rb["Name"] == name && rb["Pass"] == pass {
			u.Name = strings.Title(rb["Name"].(string))
			u.Email = rb["Email"].(string)
			return nil
		}
		return errors.New("Bad username or password.")
	}
}

// save will save a users info in users database.
func (u *user) save(name, pass, email string) error {
	if userExists(name) {
		return errors.New("User already exists.")
	}
	_, err := userDB.Insert(map[string]interface{}{
		"Name":  strings.ToLower(name),
		"Pass":  pass,
		"Email": email})
	if err != nil {
		return err
	}
	return nil
}
