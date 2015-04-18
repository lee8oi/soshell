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
	"golang.org/x/net/websocket"
	//"log"
	"strings"
)

// appendMsg appends a msg (div.msg) element to selector.
func appendMsg(ws *websocket.Conn, selector, text string) (e error) {
	p := newPacket("appendElement")
	p.Map["Element"] = "div"
	p.Map["Selector"] = selector
	p.Map["Class"] = "msg"
	p.Map["Text"] = text
	p.Map["Scroll"] = "true"
	e = sendPacket(ws, p)
	return
}

// setAttribute sets the specified attribute for selector.
func setAttribute(ws *websocket.Conn, selector, attribute, value string) (e error) {
	p := newPacket("setAttribute")
	p.Map["Selector"] = selector
	p.Map["Attribute"] = attribute
	p.Map["Value"] = value
	e = sendPacket(ws, p)
	return
}

// getAttribute returns the current value of an attribute of selector.
func getAttribute(ws *websocket.Conn, selector, attribute string) (s string, e error) {
	p := newPacket("getAttribute")
	p.Map["Selector"] = selector
	p.Map["Attribute"] = attribute
	e = sendPacket(ws, p)
	if e == nil {
		resp, e := readPacket(ws)
		if e == nil {
			if response, ok := resp.Map["Response"]; ok {
				s = response
			}
		}
	}
	return
}

// setProperty sets the specified CSS property of selector.
func setProperty(ws *websocket.Conn, selector, property, value string) (e error) {
	p := newPacket(property)
	p.Map["Selector"] = selector
	p.Map["Value"] = value
	e = sendPacket(ws, p)
	return
}

// getProperty returns the current (computed) value for the specified CSS property of selector.
func getProperty(ws *websocket.Conn, selector, property string) (s string, e error) {
	p := newPacket("getProperty")
	p.Map["Selector"] = selector
	p.Map["Property"] = property
	e = sendPacket(ws, p)
	if e == nil {
		resp, e := readPacket(ws)
		if e == nil {
			if response, ok := resp.Map["Response"]; ok {
				s = response
			}
		}
	}
	return
}

// prompt sends the specified text as a msg and returns user input as a string.
func prompt(ws *websocket.Conn, text string) (s string, e error) {
	if len(text) > 0 {
		e = appendMsg(ws, "#msgList", text)
	} else {
		e = appendMsg(ws, "#msgList", "Enter some input:")
	}
	if e == nil {
		input, e := readPacket(ws)
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
func promptSecure(ws *websocket.Conn, selector, text string) (input string, e error) {
	attr, e := getAttribute(ws, selector, "type")
	if e == nil {
		defer setAttribute(ws, selector, "type", attr)
		e = setAttribute(ws, selector, "type", "password")
		if e == nil {
			input, e = prompt(ws, text)
		}
	}
	return
}

type command struct {
	Desc    string
	Handler func(*websocket.Conn, packet) error
}

var cmdMap = make(map[string]command)

func init() {
	cmdMap["help"] = command{
		Desc: "help returns help information about available commands.",
		Handler: func(ws *websocket.Conn, pack packet) (e error) {
			if len(pack.Args) > 0 {
				if len(pack.Args) == 1 {
					cmds := ""
					for k, _ := range cmdMap {
						cmds += " " + k
					}
					e = appendMsg(ws, "#msgList", "Available commands:"+cmds)
				} else {
					if c, ok := cmdMap[pack.Args[1]]; ok {
						e = appendMsg(ws, "#msgList", c.Desc)
					} else {
						e = appendMsg(ws, "#msgList", "Command not available: "+pack.Args[1])
					}
				}
			}
			return
		},
	}
	cmdMap["prompt"] = command{
		Desc: "prompt is a testing command that prompts the user for input.",
		Handler: func(ws *websocket.Conn, pack packet) (e error) {
			if len(pack.Args) > 0 {
				text, e := prompt(ws, strings.Join(pack.Args[1:], " "))
				if e == nil && len(text) > 0 {
					e = appendMsg(ws, "#msgList", "You said: "+text)
				}
			}
			return
		},
	}
	cmdMap["login"] = command{
		Desc: "login is an experimental login command.",
		Handler: func(ws *websocket.Conn, pack packet) (e error) {
			if len(pack.Args) > 0 {
				if len(pack.Args) == 1 {
					e = appendMsg(ws, "#msgList", "Usage: login <name>")
				} else {
					name := pack.Args[1]
					e = appendMsg(ws, "#msgList", "Hello, "+name+"!")
					if e == nil {
						pass, e := promptSecure(ws, "#msgTxt", "Please enter your password")
						if e == nil && len(pass) > 0 {
							e = appendMsg(ws, "#msgList", "You entered User:"+name+" and pass:"+pass)

						}
					}
				}
			}
			return
		},
	}
}
