/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

//
package main

import (
	"flag"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
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

// packet is an extensible object type transmitted via websocket as JSON.
type packet struct {
	Type string
	Args []string
	Map  map[string]string
}

// client is an extensible type representing a single websocket client.
type client struct {
	ws            *websocket.Conn
	user          user
	path, address string
}

// checkTLS returns "SECURED" if TLS handshake is complete or "UNSECURED" if not.
func checkTLS(r *http.Request) string {
	if r.TLS != nil && r.TLS.HandshakeComplete {
		return "SECURED"
	}
	return "UNSECURED"
}

// newPacket returns an initialized packet. Any arguments are added to the pack.Args
// and the first arg is used for pack.Type.
func newPacket(args ...string) (pack packet) {
	pack.Map = make(map[string]string)
	if len(args) > 0 {
		if len(args) > 1 {
			pack.Type = args[0]
			pack.Args = append(pack.Args, args[1:]...)
		} else {
			pack.Type = args[0]
		}
	}
	return
}

// listener listens for incoming packets and passes them to the respective handlers.
func (c *client) listener() (e error) {
	for {
		var p packet
		e = c.ws.ReadJSON(&p)
		if e != nil {
			if e != io.EOF {
				log.Println(e)
			}
			break
		}
		if len(p.Args) > 0 && len(p.Args[0]) > 0 {
			if cmd, exists := cmdMap[strings.ToLower(p.Args[0])]; exists {
				e = cmd.Handler(c, p)
			} else {
				e = c.appendMsg("#msg-list", p.Args[0]+": command not found ")
			}
		} else {
			log.Println("not enough args")
		}
		time.Sleep(time.Second)
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
	var c = client{ws: ws, address: ws.RemoteAddr().String()}
	log.Println(c.address, r.URL, "connected")
	c.innerHTML("#status-box", "<b>"+checkTLS(r)+"</b>")
	e := c.listener()
	if e != nil && e != io.EOF {
		log.Println(e)
	}
	log.Println(c.address, "disconnected")
}

func serveClient(w http.ResponseWriter, r *http.Request) {
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
	clientTempl.Execute(w, data{SockUrl: sockUrl, Status: "HTTP " + checkTLS(r)})
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
	go func() {
		// cert.pem is ssl.crt + *server.ca.pem
		err := http.ListenAndServeTLS(*httpsAddr, *certFile, *keyFile, nil)
		if err != nil {
			log.Fatal("ListenAndServeTLS:", err)
		}
	}()
	err := http.ListenAndServe(*httpAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
