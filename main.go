/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"golang.org/x/net/websocket"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"
	//"strings"
)

var addr = flag.String("http", ":8080", "http service address")
var addrs = flag.String("https", ":8090", "https service address")
var hostname = flag.String("host", "localhost", "domain or host name")

// packet is an extensible object type transmitted via websocket as JSON.
type packet struct {
	Type string
	Args []string
	Map  map[string]string
}

// checkTLS returns "SECURED" if TLS handshake is complete or "UNSECURED" if not.
func checkTLS(r *http.Request) string {
	if r.TLS != nil && r.TLS.HandshakeComplete {
		return "SECURED"
	}
	return "UNSECURED"
}

/*
newPacket returns an initialized packet. Any arguments are added to the pack.Args
and the first arg is used for pack.Type.
*/
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

// readPacket reads a single packet from a websocket.
func readPacket(ws *websocket.Conn) (p packet, e error) {
	e = websocket.JSON.Receive(ws, &p)
	return
}

// sendPacket converts a packet to JSON then writes it to the websocket.
func sendPacket(ws *websocket.Conn, pack packet) (e error) {
	if j, e := json.Marshal(pack); e == nil {
		_, e = ws.Write(j)
	}
	if e != nil {
		log.Println(e)
	}
	return
}

/*
packHandler reads all incoming packets from the websocket and checks for
command handlers.
*/
func packetHandler(ws *websocket.Conn) (e error) {
	for {
		p, e := readPacket(ws)
		if e == nil {
			if len(p.Args) > 0 {
				if cmd, ok := cmdMap[p.Args[0]]; ok {
					e = cmd.Handler(ws, p)
				} else {
					e = appendMsg(ws, "#msgList", p.Args[0]+": command not found ")
				}
			} else {
				e = errors.New("Args: object missing")
			}
		}
		if e != nil {
			break
		}
		time.Sleep(time.Second)
	}
	return
}

// sockHandler handles individual websocket connections.
func sockHandler(ws *websocket.Conn) {
	if ws.Config().Origin.String() != "https://"+*hostname+*addrs {
		log.Println("Bad Origin!", ws.Config().Origin)
	} else {
		if e := appendMsg(ws, "#msgList", "SOCKET "+checkTLS(ws.Request())); e == nil {
			defer log.Println(ws.Request().RemoteAddr, "disconnected")
			log.Println(ws.Request().RemoteAddr, "connected")
			e = packetHandler(ws)
			if e != nil && e != io.EOF {
				log.Println(e)
			}
		}
	}
}

var clientTemplate = template.Must(template.ParseFiles("client.html"))

// clientServer serves the websocket client to the requesting browser.
func clientServer(w http.ResponseWriter, r *http.Request) {
	type data struct {
		SockUrl, Status string
	}
	sockUrl := "wss://" + *hostname + *addrs + "/sock"
	clientTemplate.Execute(w, data{SockUrl: sockUrl, Status: "HTTP " + checkTLS(r)})
}

func main() {
	flag.Parse()
	http.Handle("/", http.HandlerFunc(clientServer))
	http.Handle("/sock", websocket.Handler(sockHandler))
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	go func() {
		// cert.pem is ssl.crt + *server.ca.pem
		err := http.ListenAndServeTLS(*addrs, "cert.pem", "key.pem", nil)
		if err != nil {
			log.Fatal("ListenAndServeTLS:", err)
		}
	}()
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
