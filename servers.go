/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

//
package main

import (
	"errors"
	"log"
)

var servers serverList

func init() {
	servers = make(serverList)
}

type serverList map[string]*server

func (sl *serverList) exists(name string) (b bool) {
	_, b = (*sl)[name]
	return
}

func (c *client) connect(name string) {
	if servers.exists(name) == false {
		servers[name] = newServer(name)
		go servers[name].hub()
	}
	c.server = name
	servers[name].connect <- c
	servers[name].broadcast <- c.user.Name + " has connected."
	c.command = &chatCommands
}

func (c *client) disconnect() error {
	name := c.server
	if servers.exists(name) {
		(*servers[name]).broadcast <- c.user.Name + " has disconnected."
		(*servers[name]).disconnect <- c
		c.server = ""
		c.command = &sysCommands
		return nil
	}
	return errors.New("Not connected to a server.")
}

type server struct {
	connections map[string]*client
	connect     chan *client
	disconnect  chan *client
	broadcast   chan string
	name        string
}

func (s *server) empty() bool {
	if len(s.connections) == 0 {
		return true
	}
	return false
}

func (s *server) hub() {
	defer log.Println("Server closed")
	for {
		select {
		case c := <-s.connect:
			s.connections[c.user.Name] = c
		case c := <-s.disconnect:
			delete(s.connections, c.user.Name)
			if s.empty() {
				delete(servers, s.name)
				return
			}
		case msg := <-s.broadcast:
			for _, v := range s.connections {
				v.appendMsg("#msg-list", msg)
			}
		}
	}
}

func newServer(name string) (s *server) {
	s = new(server)
	s.name = name
	s.connections = make(map[string]*client)
	s.connect = make(chan *client)
	s.disconnect = make(chan *client)
	s.broadcast = make(chan string)
	return
}
