/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

/*
This file contains the command addons used used by the client via text commands.
*/

//
package main

import (
	"log"
)

type command struct {
	Desc    string
	Handler func(*client, []string) error
}

var cmdMap = make(map[string]command)

func init() {
	cmdMap["help"] = command{
		Desc: "help returns help information about available commands.",
		Handler: func(c *client, args []string) (e error) {
			if len(args) > 0 {
				if len(args) == 1 {
					cmds := ""
					for k := range cmdMap {
						cmds += " " + k
					}
					e = c.appendMsg("#msg-list", "Available commands:"+cmds)
				} else {
					if cmd, ok := cmdMap[args[1]]; ok {
						e = c.appendMsg("#msg-list", cmd.Desc)
					} else {
						e = c.appendMsg("#msg-list", "Command not available: "+args[1])
					}
				}
			}
			return
		},
	}
	//	cmdMap["motd"] = command{
	//		Desc: "motd prints the current message-of-the-day.",
	//		Handler: func(c *client, args []string) (e error) {
	//			if len(args) > 0 {
	//				// do something.
	//			}
	//			return
	//		},
	//	}
	cmdMap["clear"] = command{
		Desc: "clear the current terminal's content",
		Handler: func(c *client, args []string) (e error) {
			if len(args) > 0 {
				c.innerHTML("#msg-list", " ")
			}
			return
		},
	}
	cmdMap["login"] = command{
		Desc: "login lets you log into a registered user account.",
		Handler: func(c *client, args []string) (e error) {
			if len(args) > 0 {
				if len(args) == 1 {
					e = c.appendMsg("#msg-list", "Usage: login <name>")
				} else {
					name := args[1]
					if isName(name) {
						if userExists(name) {
							pass, e := c.promptSecure("#msg-txt", "Please enter your password")
							if e == nil && len(pass) > 0 {
								e = c.user.login(name, pass)
								if e != nil {
									log.Println("login error:", e)
									e = c.appendMsg("#msg-list", "Login failed")
								} else {
									e = c.innerHTML("#status-box", "<b>"+c.user.Name+"</b>")
									if e == nil {
										e = c.appendMsg("#msg-list", "Welcome back, "+c.user.Name)
									}
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
		Handler: func(c *client, args []string) (e error) {
			if len(args) > 1 {
				name := args[1]
				if isName(name) {
					if !userExists(name) {
						email, e := c.prompt("Enter your email address")
						if e == nil && isEmail(email) {
							pass1, e := c.promptSecure("#msg-txt", "Enter a good password")
							if e == nil {
								pass2, e := c.promptSecure("#msg-txt", "Re-enter your password")
								if e == nil && pass1 == pass2 {
									c.user.Email = email
									c.user.Name = name
									e = c.user.save(name, pass1, email)
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
						e = c.appendMsg("#msg-list", "User already exists")
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
