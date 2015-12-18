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

// listener listens for incoming packets and passes them to the input parser.
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

// parseInput splits up input and passes the arguments to the respective command handler.
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
				servers[c.server].broadcast <- fmt.Sprintf("[%s] %s", c.user.Name, string(b))
			}
		} else {
			e = errors.New("Command failed.")
		}
	}
	return
}

// runCommand ensures the command specified in the input exists then runs it.
func (c *client) runCommand(args []string) (e error) {
	if cmd, exists := (*c.command)[strings.ToLower(args[0])]; exists {
		e = cmd.Handler(c, args)
	} else {
		e = errors.New("Command not found.")
	}
	return
}

// write take a string and writes it to the websocket.
func (c *client) write(msg string) (e error) {
	e = c.ws.WriteMessage(websocket.TextMessage, []byte(msg))
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
	elem := "<div class=\"msg\">" + text + "</div>"
	c.write(fmt.Sprintf("append(\"%s\", '%s')", selector, elem))
	c.write(fmt.Sprintf("scroll(\"%s\")", selector))
	return
}

// focus will set the window focus on selector
func (c *client) focus(selector, value string) (e error) {
	str := fmt.Sprintf("focus(\"%s\", %s)", selector, value)
	e = c.write(str)
	return
}

// exists will check if selector exists
func (c *client) exists(selector string) (bl bool) {
	e := c.write(fmt.Sprintf("exists(\"%s\")", selector))
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
	e = c.write(fmt.Sprintf("innerHTML(\"%s\", '%s')", selector, value))
	return
}

// getHTML returns the innerHTML of selector
func (c *client) getHTML(selector string) (s string, e error) {
	if c.exists(selector) {
		e = c.write(fmt.Sprintf("getHTML(\"%s\")", selector))
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
	e = c.write(fmt.Sprintf("setAttribute(\"%s\", \"%s\", \"%s\")", selector, attribute, value))
	return
}

// getAttribute returns the current value of an attribute of selector.
func (c *client) getAttribute(selector, attribute string) (s string, e error) {
	e = c.write(fmt.Sprintf("getAttribute(\"%s\", \"%s\")", selector, attribute))
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
	e = c.write(fmt.Sprintf("setProperty(\"%s\", \"%s\", \"%s\")", selector, property, value))
	return
}

// getProperty returns the current (computed) value for the specified CSS property of selector.
func (c *client) getProperty(selector, property string) (s string, e error) {
	e = c.write(fmt.Sprintf("getProperty(\"%s\", \"%s\")", selector, property))
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
	e = c.write(fmt.Sprintf("editable(\"%s\", %s)", selector, value))
	return
}
