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
	"bufio"
	"errors"
	//"log"
	"regexp"
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

func (c *client) appendLink(selector, url, text string) (e error) {
	p := newPacket("appendElement")
	p.Map["Element"] = "a"
	p.Map["Selector"] = selector
	p.Map["Id"] = text
	p.Map["Class"] = "ip-link"
	p.Map["Href"] = url
	p.Map["Text"] = text
	p.Map["Target"] = "_blank"
	p.Map["Scroll"] = "true"
	p.Map["OnClick"] = "removeDecoration"
	e = c.ws.WriteJSON(p)
	return
}

func (c *client) appendBreak(selector string) (e error) {
	p := newPacket("appendElement")
	p.Map["Element"] = "br"
	p.Map["Selector"] = selector
	p.Map["Scroll"] = "true"
	e = c.ws.WriteJSON(p)
	return
}

// focus will set the window focus on selector
func (c *client) focus(selector, value string) (e error) {
	p := newPacket("focus")
	p.Map["Selector"] = selector
	p.Map["Value"] = value
	e = c.ws.WriteJSON(p)
	return
}

// exists will check if selector exists
func (c *client) exists(selector string) (b bool) {
	p := newPacket("exists")
	p.Map["Selector"] = selector
	e := c.ws.WriteJSON(p)
	if e == nil {
		var resp packet
		e = c.ws.ReadJSON(&resp)
		if e == nil {
			if response, ok := resp.Map["Response"]; ok {
				if response == "true" {
					return true
				} else {
					return false
				}
			}
		}
	}
	return
}

// innerHTML will set the html content of selector
func (c *client) innerHTML(selector, value string) (e error) {
	p := newPacket("innerHTML")
	p.Map["Selector"] = selector
	p.Map["Value"] = value
	e = c.ws.WriteJSON(p)
	return
}

// getHTML returns the innerHTML of selector
func (c *client) getHTML(selector string) (s string, e error) {
	if c.exists(selector) {
		p := newPacket("getHTML")
		p.Map["Selector"] = selector
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
	} else {
		e = errors.New("element does not exist")
	}
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

// editable sets the editable property of the element
func (c *client) editable(selector, value string) (e error) {
	p := newPacket("editable")
	p.Map["Selector"] = selector
	p.Map["Value"] = value
	e = c.ws.WriteJSON(p)
	return
}

// prompt sends the specified text as a msg and returns user input as a string.
func (c *client) prompt(text string) (s string, e error) {
	if len(text) > 0 {
		e = c.appendMsg("#msg-list", text)
	} else {
		e = c.appendMsg("#msg-list", "Enter some input:")
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
					e = c.appendMsg("#msg-list", "Available commands:"+cmds)
				} else {
					if cmd, ok := cmdMap[pack.Args[1]]; ok {
						e = c.appendMsg("#msg-list", cmd.Desc)
					} else {
						e = c.appendMsg("#msg-list", "Command not available: "+pack.Args[1])
					}
				}
			}
			return
		},
	}
	//	cmdMap["prompt"] = command{
	//		Desc: "prompt is a testing command that prompts the user for input.",
	//		Handler: func(c *client, pack packet) (e error) {
	//			if len(pack.Args) > 0 {
	//				text, e := c.prompt(strings.Join(pack.Args[1:], " "))
	//				if e == nil && len(text) > 0 {
	//					e = c.appendMsg("#msg-list", "You said: "+text)
	//				} else {
	//					log.Println(e)
	//				}
	//			}
	//			return
	//		},
	//	}
	cmdMap["clear"] = command{
		Desc: "clear the current terminal's content",
		Handler: func(c *client, pack packet) (e error) {
			if len(pack.Args) > 0 {
				c.innerHTML("#msg-list", " ")
			}
			return
		},
	}
	cmdMap["editor"] = command{
		Desc: "editor opens a simple editable box in the terminal",
		Handler: func(c *client, pack packet) (e error) {
			if len(pack.Args) > 0 {
				c.appendMsg("#msg-list", "Editor box:")
				p := newPacket("appendElement")
				p.Map["Element"] = "div"
				p.Map["Selector"] = "#msg-list"
				p.Map["Id"] = "editor"
				p.Map["Focus"] = "true"
				e = c.ws.WriteJSON(p)
				c.editable("#msg-list #editor", "true")
				c.setProperty("#msg-list #editor", "border", "1px solid #fff")
				c.focus("#msg-list #editor", "true")
			}
			return
		},
	}
	cmdMap["ipscraper"] = command{
		Desc: "scrapes unique ip addresses from editor box text",
		Handler: func(c *client, pack packet) (e error) {
			if len(pack.Args) > 0 {
				if c.exists("#msg-list #editor") {
					data, e := c.getHTML("#msg-list #editor")
					if e == nil {
						ipMap := make(map[string]bool)
						re := regexp.MustCompile("\\b\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\b")
						r := bufio.NewReader(strings.NewReader(data))
						for {
							line, _, err := r.ReadLine()
							if err == nil {
								items := re.FindAllStringSubmatch(string(line), -1)
								for _, val := range items {
									if _, exists := ipMap[val[0]]; exists == false {
										c.appendLink("#msg-list", "http://hackerexperience.com/internet?ip="+val[0], val[0])
										c.appendBreak("#msg-list")
										ipMap[val[0]] = true
									}
								}
							} else {
								break
							}
						}
					}
				} else {
					c.appendMsg("#msg-list", "You do not have an editor box open")
				}
			}
			return
		},
	}
	cmdMap["login"] = command{
		Desc: "login lets you log into a registered user account.",
		Handler: func(c *client, pack packet) (e error) {
			if len(pack.Args) > 0 {
				if len(pack.Args) == 1 {
					e = c.appendMsg("#msg-list", "Usage: login <name>")
				} else {
					name := pack.Args[1]
					if isName(name) {
						path := *users + SEP + indexPath([]byte(name))
						if pathExists(path) {
							pass, e := c.promptSecure("#msg-txt", "Please enter your password")
							if e == nil && len(pass) > 0 {
								e = c.user.load(name, pass)
								if e != nil {
									e = c.appendMsg("#msg-list", "Login failed")
								} else {
									e = c.appendMsg("#msg-list", "Welcome back, "+c.user.Name)
								}
							}
						} else {
							e = c.appendMsg("#msg-list", "User does not exist")
						}
					} else {
						e = c.appendMsg("#msg-list", "Invalid characters in name")
					}
				}
			}
			return
		},
	}
	cmdMap["register"] = command{
		Desc: "register a user account",
		Handler: func(c *client, pack packet) (e error) {
			if len(pack.Args) > 1 {
				name := pack.Args[1]
				if isName(name) {
					email, e := c.prompt("Enter your email address")
					if e == nil && isEmail(email) {
						pass1, e1 := c.promptSecure("#msg-txt", "Enter a good password")
						if e1 == nil {
							pass2, e2 := c.promptSecure("#msg-txt", "Re-enter your password")
							if e2 == nil && pass1 == pass2 {
								c.user.Email = email
								c.user.Name = name
								e = c.user.save(name, pass1)
								if e == nil {
									e = c.appendMsg("#msg-list", "User account created (don't forget your password!)")
								} else {
									e = c.appendMsg("#msg-list", e.Error())
								}
							} else {
								e = c.appendMsg("#msg-list", "Failed! Passwords did not match")
							}
						}
					} else {
						e = c.appendMsg("#msg-list", "Bad email address")
					}
				} else {
					e = c.appendMsg("#msg-list", "Invalid characters in name")
				}
			} else {
				e = c.appendMsg("#msg-list", "Usage: register <name>")
			}
			return
		},
	}
}
