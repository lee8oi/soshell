/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

//
package main

import (
	"errors"
	"log"
	//"strings"
)

var servers serverList

func init() {
	servers = make(serverList)
}

type serverList map[string]*server

// exists checks if the specified server is already loaded.
func (sl *serverList) exists(name string) (b bool) {
	_, b = (*sl)[name]
	return
}

// connect connects the user to the specified server and switches command handler/prefix.
func (c *client) connect(name string) {
	if servers.exists(name) == false {
		servers[name] = newServer(name)
		go servers[name].hub()
	}
	c.server = name
	servers[name].connect <- c
	servers[name].broadcast <- c.user.Name + " has connected."
	c.command = &chatCommands
	c.cmdPrefix = "/"
}

// disconnect disconnects the user from the specified server and resets command handler/prefix.
func (c *client) disconnect() error {
	name := c.server
	if servers.exists(name) && servers[name].isConnected(c.user) {
		(*servers[name]).broadcast <- c.user.Name + " has disconnected."
		(*servers[name]).disconnect <- c
		c.server = ""
		c.command = &sysCommands
		c.cmdPrefix = ""
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

// empty checks if the server is empty (no users left).
func (s *server) empty() bool {
	if len(s.connections) == 0 {
		return true
	}
	return false
}

// isConnected checks if a user is connected to the server.
func (s *server) isConnected(u user) bool {
	if _, ok := s.connections[u.Name]; ok {
		return true
	}
	return false
}

// hub starts up the appropriate channels and listens for connect's, disconnect's, and broadcast's.
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

// new Server initializes a new server with the necessary channels.
func newServer(name string) (s *server) {
	s = new(server)
	s.name = name
	s.connections = make(map[string]*client)
	s.connect = make(chan *client)
	s.disconnect = make(chan *client)
	s.broadcast = make(chan string)
	return
}
