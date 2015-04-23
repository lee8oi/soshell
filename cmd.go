/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

/*
This file contains the command addons used by the web client. Includes
the available DOM related methods used in conjunction with client-side JS scripts for
interacting with the client HTML/CSS.
*/

//
package main

import (
	"log"
	"strings"
)

// appendMsg appends a msg (div.msg) element to selector.
func (c *client) appendMsg(selector, text string) (e error) {
	p := newPacket("appendElement")
	p.Map["Element"] = "div"
	p.Map["Selector"] = selector
	p.Map["Class"] = "msg"
	p.Map["Text"] = text
	p.Map["Scroll"] = "true"
	e = c.ws.WriteJSON(p)
	return
}

// setAttribute sets the specified attribute for selector.
func (c *client) setAttribute(selector, attribute, value string) (e error) {
	p := newPacket("setAttribute")
	p.Map["Selector"] = selector
	p.Map["Attribute"] = attribute
	p.Map["Value"] = value
	e = c.ws.WriteJSON(p)
	return
}

// getAttribute returns the current value of an attribute of selector.
func (c *client) getAttribute(selector, attribute string) (s string, e error) {
	p := newPacket("getAttribute")
	p.Map["Selector"] = selector
	p.Map["Attribute"] = attribute
	e = c.ws.WriteJSON(p)
	if e == nil {
		var resp packet
		e = c.ws.ReadJSON(&resp)
		if e == nil {
			if response, ok := resp.Map["Response"]; ok {
				s = response
			}
		}
	}
	return
}

// setProperty sets the specified CSS property of selector.
func (c *client) setProperty(selector, property, value string) (e error) {
	p := newPacket(property)
	p.Map["Selector"] = selector
	p.Map["Value"] = value
	e = c.ws.WriteJSON(p)
	return
}

// getProperty returns the current (computed) value for the specified CSS property of selector.
func (c *client) getProperty(selector, property string) (s string, e error) {
	p := newPacket("getProperty")
	p.Map["Selector"] = selector
	p.Map["Property"] = property
	e = c.ws.WriteJSON(p)
	if e == nil {
		var resp packet
		e = c.ws.ReadJSON(&resp)
		if e == nil {
			if response, ok := resp.Map["Response"]; ok {
				s = response
			}
		}
	}
	return
}

// prompt sends the specified text as a msg and returns user input as a string.
func (c *client) prompt(text string) (s string, e error) {
	if len(text) > 0 {
		e = c.appendMsg("#msgList", text)
	} else {
		e = c.appendMsg("#msgList", "Enter some input:")
	}
	if e == nil {
		var input packet
		e = c.ws.ReadJSON(&input)
		if e == nil {
			if len(input.Args) > 0 {
				if len(input.Args[0]) > 0 {
					s = strings.Join(input.Args, " ")
				}
			}
		}
	}
	return
}

// promptSecure uses prompt() but changes the selector/input box type to & from password for security.
func (c *client) promptSecure(selector, text string) (input string, e error) {
	attr, e := c.getAttribute(selector, "type")
	if e == nil {
		defer c.setAttribute(selector, "type", attr)
		e = c.setAttribute(selector, "type", "password")
		if e == nil {
			input, e = c.prompt(text)
		}
	}
	return
}

type command struct {
	Desc    string
	Handler func(*client, packet) error
}

var cmdMap = make(map[string]command)

func init() {
	cmdMap["help"] = command{
		Desc: "help returns help information about available commands.",
		Handler: func(c *client, pack packet) (e error) {
			if len(pack.Args) > 0 {
				if len(pack.Args) == 1 {
					cmds := ""
					for k, _ := range cmdMap {
						cmds += " " + k
					}
					e = c.appendMsg("#msgList", "Available commands:"+cmds)
				} else {
					if cmd, ok := cmdMap[pack.Args[1]]; ok {
						e = c.appendMsg("#msgList", cmd.Desc)
					} else {
						e = c.appendMsg("#msgList", "Command not available: "+pack.Args[1])
					}
				}
			}
			return
		},
	}
	cmdMap["prompt"] = command{
		Desc: "prompt is a testing command that prompts the user for input.",
		Handler: func(c *client, pack packet) (e error) {
			if len(pack.Args) > 0 {
				text, e := c.prompt(strings.Join(pack.Args[1:], " "))
				if e == nil && len(text) > 0 {
					e = c.appendMsg("#msgList", "You said: "+text)
				} else {
					log.Println(e)
				}
			}
			return
		},
	}
	cmdMap["login"] = command{
		Desc: "login is an experimental login command.",
		Handler: func(c *client, pack packet) (e error) {
			if len(pack.Args) > 0 {
				if len(pack.Args) == 1 {
					e = c.appendMsg("#msgList", "Usage: login <name>")
				} else {
					name := pack.Args[1]
					e = c.appendMsg("#msgList", "Hello, "+name+"!")
					if e == nil {
						pass, e := c.promptSecure("#msgTxt", "Please enter your password")
						if e == nil && len(pass) > 0 {
							e = c.appendMsg("#msgList", "You entered User:"+name+" and pass:"+pass)

						}
					}
				}
			}
			return
		},
	}
}
