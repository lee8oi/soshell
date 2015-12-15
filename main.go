/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

//
package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"text/template"
)

const SEP = string(os.PathSeparator)

var (
	httpAddr    = flag.String("http", ":8080", "http service address")
	httpsAddr   = flag.String("https", ":8090", "https service address")
	hostname    = flag.String("host", "localhost", "domain or host name")
	work        = flag.String("work", "work", "working directory")
	users       = flag.String("users", "users", "users root folder")
	certFile    = flag.String("cert", "cert.pem", "SSL certificate file")
	keyFile     = flag.String("key", "key.pem", "SSL key file")
	public      = flag.String("public", "public", "public web directory")
	clientTempl *template.Template
)

// isTLS checks for TLS and returns true if handshake is complete or false if not.
func isTLS(r *http.Request) bool {
	if r.TLS != nil && r.TLS.HandshakeComplete {
		return true
	}
	return false
}

// getArgs splits a slice of bytes into a slice of string arguments.
// Anything in '', "", or `` are consider a single argument (including spaces).
func getArgs(b []byte) (s []string) {
	re := regexp.MustCompile("`([\\S\\s]*)`|('([\\S \\t\\r]*)'|\"([\\S ]*)\"|\\S+)")
	args := re.FindAllSubmatch(b, -1)
	for _, val := range args {
		s = append(s, string(val[0]))
	}
	return
}

// serveWs serves the websocket and starts the listener on successful connection.
func serveWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if r.Header.Get("Origin") != "https://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()
	var c = client{ws: ws, address: ws.RemoteAddr().String(), user: user{Name: "Guest"}}
	log.Println(c.address, r.URL, "connected")
	c.innerHTML("#status-box", "<b>"+c.user.Name+"</b>")
	e := c.listener()
	if e != nil && e != io.EOF {
		log.Println(e)
	}
	log.Println(c.address, "disconnected")
}

// serveClient is the handler that serves the client html on initial connection.
func serveClient(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RemoteAddr, r.Referer(), r.URL, "connecting")
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.TLS == nil {
		log.Println("redirecting")
		http.Redirect(w, r, "https://"+*hostname+*httpsAddr, 301)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method nod allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	type data struct {
		SockUrl, Status string
	}
	sockUrl := "wss://" + *hostname + *httpsAddr + "/ws"
	clientTempl.Execute(w, data{SockUrl: sockUrl})
}

// pathExists returns true if the path exists or false if it doesn't.
func pathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(err)
	}
	return true
}

func init() {
	flag.Parse()
	dirs := map[string]os.FileMode{*work: 0700, *public: 0755, *users: 0700}
	for path, perm := range dirs {
		if pathExists(path) {
			err := os.Chmod(path, perm)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			err := os.Mkdir(path, perm)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	for _, file := range []*string{certFile, keyFile} {
		_, err := os.Stat(*file)
		if os.IsNotExist(err) {
			path := *work + SEP + *file
			_, err := os.Stat(path)
			if err != nil {
				log.Fatal(err)
			}
			*file = path
		}
	}
	clientTempl = template.Must(template.ParseFiles(*public + SEP + "client.html"))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", serveClient)
	r.HandleFunc("/ws", serveWs)
	http.Handle("/", r)
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(*public))))
	loadUserDB()
	go func() {
		// cert.pem is ssl.crt + *server.ca.pem
		fmt.Println("Listening at " + "https://" + *hostname + *httpsAddr)
		err := http.ListenAndServeTLS(*httpsAddr, *certFile, *keyFile, nil)
		if err != nil {
			log.Fatal("ListenAndServeTLS:", err)
		}
	}()
	go func() {
		err := http.ListenAndServe(*httpAddr, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	s := <-c
	fmt.Printf("Caught %s signal. Shutting down.\n", s)
	closeUserDB()
}
