/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

/* 
This file contains the websocket functions along with the DomMap that is used inconjunction
with server-side methods to provide interactive access to client-side html/css.
*/

var prefix = "/";
var ws = new WebSocket(sockUrl);
ws.onopen = function (event) {
	document.getElementById("msgTxt").focus();
};
ws.onclose = function(){
	AppendMsg("#msgList", "Disconnected");
};
ws.onmessage = function(event) {
	var obj = JSON.parse(event.data);
	if (obj && obj["Type"]) {
		if (DomMap[obj["Type"]]) {
			RunDom(obj);
		}
	}
};
function AppendMsg(selector, text) {
	var obj = {};
	obj["Type"] = "appendElement";
	obj["Map"] = {};
	obj["Map"]["Element"] = "div";
	obj["Map"]["Selector"] = selector;
	obj["Map"]["Class"] = "msg";
	obj["Map"]["Text"] = text;
	obj["Map"]["Scroll"] = "true";
	RunDom(obj);
}
function GetArgs(str) {
	var re = /`([\S\s]*)`|('([\S \t\r]*)'|"([\S ]*)"|\S+)/g;
	return str.match(re);
}
function parseInput(text) {
	jsonStr = "";
	args = GetArgs(text);
	if (args.length > 0) {
		jsonStr = {"Type": "CMD", "Args": args};
	}
	return jsonStr;
}
function Send() {
	var obj = parseInput(document.getElementById("msgTxt").value);
	json = JSON.stringify(obj);
	document.getElementById("msgTxt").value = "";
	ws.send(json);
	return false
}
function Respond(str) {
	var resp = {};
	resp["Type"] = "RESP";
	resp["Map"] = {};
	resp["Map"]["Response"] = str;
	json = JSON.stringify(resp);
	ws.send(json)
}
function RunDom(obj) {
	if (obj && obj["Map"]["Selector"]) {
		var elem = document.querySelector(obj["Map"]["Selector"]);
		if (elem && obj["Type"] && obj["Type"].length > 0) {
			DomMap[obj["Type"]](elem, obj);
		}
	}
}
var DomMap = {};
DomMap["appendElement"] = function (elem, obj) {
	if (obj["Map"]["Element"]) {
		var node = document.createElement(obj["Map"]["Element"]);
		if (obj["Map"]["Class"]) {
			node.className = obj["Map"]["Class"];
		}
		if (obj["Map"]["Id"]) {
			node.id = obj["Map"]["Id"];
		}
		if (obj["Map"]["Attribute"] && obj["Map"]["Value"]) {
			node.setAttribute(obj["Map"]["Attribute"], obj["Map"]["Value"]);
		}
		if (obj["Map"]["Text"]) {
			var text = document.createTextNode(obj["Map"]["Text"]);
	   		node.appendChild(text);
		}
   		elem.appendChild(node);
		if (obj["Map"]["Scroll"] && obj["Map"]["Scroll"] == "true") {
			elem.scrollTop = elem.scrollHeight;
		}
	}
}
DomMap["setAttribute"] = function (elem, obj) {
	if (obj["Map"]["Attribute"] && obj["Map"]["Value"]) {
		elem.setAttribute(obj["Map"]["Attribute"], obj["Map"]["Value"]);
	}
}
DomMap["getAttribute"] = function (elem, obj) {
	if (obj["Map"]["Attribute"]) {
		Respond(elem.getAttribute(obj["Map"]["Attribute"]));
	}
}
DomMap["getProperty"] = function (elem, obj) {
	if (obj["Map"]["Property"]) {
		Respond(window.getComputedStyle(elem,null).getPropertyValue(obj["Map"]["Property"]));
	}
}
DomMap["background"] = function (elem, obj) {
	if (obj["Map"]["Value"]) {
		elem.style.background = obj["Map"]["Value"];
	}
}
DomMap["background-color"] = function (elem, obj) {
	if (obj["Map"]["Value"]) {
		elem.style.backgroundColor = obj["Map"]["Value"];
	}
}
DomMap["color"] = function (elem, obj) {
	if (obj["Map"]["Value"]) {
		elem.style.color = obj["Map"]["Value"];
	}
}
DomMap["border"] = function (elem, obj) {
	if (obj["Map"]["Value"]) {
		elem.style.border = obj["Map"]["Value"];
	}
}
DomMap["border-color"] = function (elem, obj) {
	if (obj["Map"]["Value"]) {
		elem.style.borderColor = obj["Map"]["Value"];
	}
}