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
	// "github.com/HouzuoGuo/tiedot/dberr"
)

var userDB *db.Col

func init() {
	database, err := db.OpenDB(*work + SEP + "database")
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Create("users"); err == nil {
		if err := userDB.Index([]string{"Name"}); err != nil {
			log.Println(err)
		}
		if err := userDB.Index([]string{"Pass"}); err != nil {
			log.Println(err)
		}
		if err := userDB.Index([]string{"Email"}); err != nil {
			log.Println(err)
		}
	}
	userDB = database.Use("users")
}

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

// load is used to load a users info from the users database.
func (u *user) load(name, pass string) error {
	var query interface{}
	json.Unmarshal([]byte(`[{"eq": "`+name+`", "in": ["Name"]}]`), &query)
	result := make(map[int]struct{})
	if err := db.EvalQuery(query, userDB, &result); err != nil {
		log.Println(err)
		return err
	}
	var (
		rb  map[string]interface{}
		err error
	)
	for id := range result {
		rb, err = userDB.Read(id)
		if err != nil {
			return err
		}
		break //only need one result
	}
	if rb["Name"] == name && rb["Pass"] == pass {
		u.Name = rb["Name"].(string)
		return nil
	}
	return errors.New("Bad username or password.")
}

// save will save a users info in users database.
func (u *user) save(name, pass string) error {
	_, err := userDB.Insert(map[string]interface{}{
		"Name":  name,
		"Pass":  pass,
		"Email": "blank"})
	if err != nil {
		return err
	}
	return nil
}
