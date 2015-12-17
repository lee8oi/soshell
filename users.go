/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

//
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/HouzuoGuo/tiedot/db"
	// "github.com/HouzuoGuo/tiedot/dberr"
)

var (
	database *db.DB
	userDB   *db.Col
)

var (
	users     map[string]*user
	guestlist map[string]bool
)

type user struct {
	Email, Name string
	auth        bool
	ID          int
}

func init() {
	users = make(map[string]*user)
	guestlist = make(map[string]bool)
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

func randNum() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ar := r.Perm(5)
	return fmt.Sprintf("%d%d%d%d%d", ar[0], ar[1], ar[2], ar[3], ar[4])
}

func guestName() string {
	name := "Guest" + randNum()
	for {
		if _, ok := guestlist[name]; ok {
			name = "Guest" + randNum()
		} else {
			guestlist[name] = true
			break
		}
	}
	return name
}

func loadUserDB() {
	var err error
	database, err = db.OpenDB(*dbpath + SEP + "database")
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
	_, rb, err := queryUser(name)
	if err != nil {
		log.Println(err)
	}
	if rb["Name"] == strings.ToLower(name) {
		return true
	}
	return false
}

func userID(name string) int {
	var query interface{}
	result := make(map[int]struct{})
	json.Unmarshal([]byte(`[{"eq": "`+strings.ToLower(name)+`", "in": ["Name"]}]`), &query)
	if err := db.EvalQuery(query, userDB, &result); err != nil {
		log.Println(err)
		return 0
	}
	for id := range result {
		return id
		break //only need one result
	}
	return 0
}

func userDoc(id int) (rb map[string]interface{}, err error) {
	rb, err = userDB.Read(id)
	if err != nil {
		return nil, err
	}
	return
}

func queryUser(name string) (id int, rb map[string]interface{}, err error) {
	if id := userID(name); id == 0 {
		return 0, nil, errors.New("User not found.")
	} else {
		rb, err = userDoc(id)
		if err != nil {
			return 0, nil, err
		}
	}
	return
}

// login checks the users password and loads their info from the users database.
func (u *user) login(name, pass string) error {
	if id, doc, err := queryUser(name); err != nil {
		log.Println(err)
		return err
	} else {
		name = strings.ToLower(name)
		if doc["Name"] == name && doc["Pass"] == pass {
			u.Name = strings.Title(doc["Name"].(string))
			u.Email = doc["Email"].(string)
			u.ID = id
			u.auth = true
			users[name] = u
			return nil
		}
		return errors.New("Bad username or password.")
	}
}

func (u *user) logout() error {
	if u.auth == true {
		fmt.Println("authed")
		delete(users, strings.ToLower(u.Name))
		u.Name = guestName()
		u.Email = "blank"
		u.auth = false
		u.ID = 0
		return nil
	}
	return errors.New("Not logged in.")
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
