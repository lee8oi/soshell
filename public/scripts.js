/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

/* 
This file contains the websocket functions along with the DomMap that is used inconjunction
with server-side methods to provide interactive access to client-side html/css.
*/

var prefix = "/";
var ws
function startSock() {
	ws = new WebSocket(sockUrl);
	ws.onopen = function (event) {
		AppendMsg("#msg-list", "Connected");
		document.getElementById("msg-txt").focus();
	};
	ws.onclose = function(){
		AppendMsg("#msg-list", "Disconnected");
		setTimeout(startSock, 3000);
	};
	ws.onmessage = function(event) {
		var obj = JSON.parse(event.data);
		if (obj && obj["Type"]) {
			if (DomMap[obj["Type"]]) {
				RunDom(obj);
			}
		}
	};
}
startSock();
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
function Send() {
	var elem = document.getElementById("msg-txt")
	ws.send(elem.value);
	elem.value = "";
	return false
}
var OnClick = {};
OnClick["removeDecoration"] = function (obj) {
	obj.onclick = function() {
		obj.style.textDecoration = "none";
	}
}
function RunDom(obj) {
	if (obj && obj["Map"]["Selector"]) {
		var elem = document.querySelector(obj["Map"]["Selector"]);
		if (obj["Type"] && obj["Type"].length > 0) {
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
		if (obj["Map"]["HTML"]) {
			node.innerHTML = obj["Map"]["HTML"];
		}
		if (obj["Map"]["Href"]) {
			node.href = obj["Map"]["Href"];
		}
		if (obj["Map"]["Target"]) {
			node.target = obj["Map"]["Target"];
		}
		if (obj["Map"]["OnClick"] && OnClick[obj["Map"]["OnClick"]]) {
			OnClick[obj["Map"]["OnClick"]](node);
		}
		if (obj["Map"]["Focus"] === "true") {
			elem.focus();
		}
   		elem.appendChild(node);
		if (obj["Map"]["Scroll"] && obj["Map"]["Scroll"] == "true") {
			elem.scrollTop = elem.scrollHeight;
		}
	}
}
DomMap["innerHTML"] = function (elem, obj) {
	if (obj["Map"]["Value"]) {
		elem.innerHTML = obj["Map"]["Value"];
	}
}
DomMap["editable"] = function (elem, obj) {
	if (obj["Map"]["Value"]) {
		if (obj["Map"]["Value"] === "true") {
			elem.contentEditable = true;
		} else {
			elem.contentEditable = false;
		}
	}
}
DomMap["focus"] = function (elem, obj) {
	if (obj["Map"]["Value"]) {
		if (obj["Map"]["Value"] === "true") {
			elem.focus();
		} else {
			elem.blur();
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
		ws.send(elem.getAttribute(obj["Map"]["Attribute"]));
	}
}
DomMap["getProperty"] = function (elem, obj) {
	if (obj["Map"]["Property"]) {
		ws.send(window.getComputedStyle(elem,null).getPropertyValue(obj["Map"]["Property"]));
	}
}
DomMap["exists"] = function (elem, obj) {
	if (elem) { 
		ws.send("true")
	} else {
		ws.send("false")
	}
}
DomMap["getHTML"] = function (elem, obj) {
	ws.send(elem.innerHTML);
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