/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

/*
The client object represents a single websocket client. It includes methods for
sending & recieving messages as well as methods for interacting with clientside
HTML & CSS via JavaScript.
*/

//
package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// packet is an extensible object type transmitted via websocket as JSON.
type packet struct {
	Type string
	Data map[string]string
}

// newPacket returns an initialized packet with Type set to t
func newPacket(t string) (pack packet) {
	pack.Data = make(map[string]string)
	pack.Type = t
	return
}

// client is an extensible type representing a single websocket client.
type client struct {
	ws            *websocket.Conn
	user          user
	path, address string
	server        string
	command       *map[string]command
	cmdPrefix     string
}

// recieve reads a single message and returns it.
func (c *client) recieve() (b []byte, e error) {
	t, m, e := c.ws.ReadMessage()
	if t == 1 {
		b = m
	}
	return
}

// listener listens for incoming packets and passes them to the respective handlers.
func (c *client) listener() (e error) {
	for {
		b, e := c.recieve()
		if e != nil {
			return e
		}
		e = c.parseInput(b)
		if e != nil {
			e = c.appendMsg("#msg-list", e.Error())
		}
		time.Sleep(time.Second)
	}
	return
}

func (c *client) parseInput(b []byte) (e error) {
	args := getArgs(b)
	if len(args) > 0 && len(args[0]) > 0 {
		if c.cmdPrefix != "" && strings.Index(args[0], c.cmdPrefix) == 0 && len(args[0]) > 1 {
			args[0] = strings.SplitN(args[0], c.cmdPrefix, 2)[1]
			e = c.runCommand(args)
		} else if c.cmdPrefix == "" {
			e = c.runCommand(args)
		} else if c.server != "" {
			if servers.exists(c.server) {
				servers[c.server].broadcast <- fmt.Sprintf("<%s> %s", c.user.Name, string(b))
			}
		} else {
			e = errors.New("Command failed.")
		}
	}
	return
}

func (c *client) runCommand(args []string) (e error) {
	if cmd, exists := (*c.command)[strings.ToLower(args[0])]; exists {
		e = cmd.Handler(c, args)
	} else {
		e = errors.New("Command not found.")
	}
	return
}

// prompt sends the specified text as a msg and returns user input as a string.
func (c *client) prompt(text string) (s string, e error) {
	if len(text) > 0 {
		e = c.appendMsg("#msg-list", text)
	} else {
		e = c.appendMsg("#msg-list", "Enter some input:")
	}
	b, e := c.recieve()
	if e == nil {
		s = string(b)
	}
	return
}

// promptSecure uses prompt() but changes the selector/input box type to & from password for security.
func (c *client) promptSecure(selector, text string) (s string, e error) {
	attr, e := c.getAttribute(selector, "type")
	if e == nil {
		defer c.setAttribute(selector, "type", attr)
		e = c.setAttribute(selector, "type", "password")
		if e == nil {
			s, e = c.prompt(text)
		}
	}
	return
}

// appendMsg appends a msg (div.msg) element to selector.
func (c *client) appendMsg(selector, text string) (e error) {
	p := newPacket("appendElement")
	p.Data["Element"] = "div"
	p.Data["Selector"] = selector
	p.Data["Class"] = "msg"
	p.Data["Text"] = text
	p.Data["Scroll"] = "true"
	e = c.ws.WriteJSON(p)
	return
}

func (c *client) appendLink(selector, url, text string) (e error) {
	p := newPacket("appendElement")
	p.Data["Element"] = "a"
	p.Data["Selector"] = selector
	p.Data["Id"] = text
	p.Data["Class"] = "ip-link"
	p.Data["Href"] = url
	p.Data["Text"] = text
	p.Data["Target"] = "_blank"
	p.Data["Scroll"] = "true"
	p.Data["OnClick"] = "removeDecoration"
	e = c.ws.WriteJSON(p)
	return
}

func (c *client) appendBreak(selector string) (e error) {
	p := newPacket("appendElement")
	p.Data["Element"] = "br"
	p.Data["Selector"] = selector
	p.Data["Scroll"] = "true"
	e = c.ws.WriteJSON(p)
	return
}

// focus will set the window focus on selector
func (c *client) focus(selector, value string) (e error) {
	p := newPacket("focus")
	p.Data["Selector"] = selector
	p.Data["Value"] = value
	e = c.ws.WriteJSON(p)
	return
}

// exists will check if selector exists
func (c *client) exists(selector string) (bl bool) {
	p := newPacket("exists")
	p.Data["Selector"] = selector
	e := c.ws.WriteJSON(p)
	if e == nil {
		b, e := c.recieve()
		if e == nil && string(b) == "true" {
			return true
		}
	}
	return false
}

// innerHTML will set the html content of selector
func (c *client) innerHTML(selector, value string) (e error) {
	p := newPacket("innerHTML")
	p.Data["Selector"] = selector
	p.Data["Value"] = value
	e = c.ws.WriteJSON(p)
	return
}

// getHTML returns the innerHTML of selector
func (c *client) getHTML(selector string) (s string, e error) {
	if c.exists(selector) {
		p := newPacket("getHTML")
		p.Data["Selector"] = selector
		e = c.ws.WriteJSON(p)
		if e == nil {
			b, e := c.recieve()
			if e == nil {
				s = string(b)
			}
		}
	} else {
		e = errors.New("element does not exist")
	}
	return
}

// setAttribute sets the specified attribute for selector.
func (c *client) setAttribute(selector, attribute, value string) (e error) {
	p := newPacket("setAttribute")
	p.Data["Selector"] = selector
	p.Data["Attribute"] = attribute
	p.Data["Value"] = value
	e = c.ws.WriteJSON(p)
	return
}

// getAttribute returns the current value of an attribute of selector.
func (c *client) getAttribute(selector, attribute string) (s string, e error) {
	p := newPacket("getAttribute")
	p.Data["Selector"] = selector
	p.Data["Attribute"] = attribute
	e = c.ws.WriteJSON(p)
	if e == nil {
		b, e := c.recieve()
		if e == nil {
			s = string(b)
		}
	}
	return
}

// setProperty sets the specified CSS property of selector.
func (c *client) setProperty(selector, property, value string) (e error) {
	p := newPacket(property)
	p.Data["Selector"] = selector
	p.Data["Value"] = value
	e = c.ws.WriteJSON(p)
	return
}

// getProperty returns the current (computed) value for the specified CSS property of selector.
func (c *client) getProperty(selector, property string) (s string, e error) {
	p := newPacket("getProperty")
	p.Data["Selector"] = selector
	p.Data["Property"] = property
	e = c.ws.WriteJSON(p)
	if e == nil {
		b, e := c.recieve()
		if e == nil {
			s = string(b)
		}
	}
	return
}

// editable sets the editable property of the element
func (c *client) editable(selector, value string) (e error) {
	p := newPacket("editable")
	p.Data["Selector"] = selector
	p.Data["Value"] = value
	e = c.ws.WriteJSON(p)
	return
}
