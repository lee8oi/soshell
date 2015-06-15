/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

/* 
This file contains the websocket functions along with the DomMap that is used inconjunction
with server-side methods to provide interactive access to client-side html/css.
*/

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
	obj.Data = {};
	obj.Data.Element = "div";
	obj.Data.Selector = selector;
	obj.Data.Class = "msg";
	obj.Data.Text = text;
	obj.Data.Scroll = "true";
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
	if (obj && obj.Data.Selector) {
		var elem = document.querySelector(obj.Data.Selector);
		if (obj.Type && obj.Type.length > 0) {
			DomMap[obj.Type](elem, obj);
		}
	}
}
var DomMap = {};
DomMap["appendElement"] = function (elem, obj) {
	if (obj.Data.Element) {
		var node = document.createElement(obj.Data.Element);
		if (obj.Data.Class) {
			node.className = obj.Data.Class;
		}
		if (obj.Data.Id) {
			node.id = obj.Data.Id;
		}
		if (obj.Data.Attribute && obj.Data.Value) {
			node.setAttribute(obj.Data.Attribute, obj.Data.Value);
		}
		if (obj.Data.Text) {
			var text = document.createTextNode(obj.Data.Text);
	   		node.appendChild(text);
		}
		if (obj.Data.HTML) {
			node.innerHTML = obj.Data.HTML;
		}
		if (obj.Data.Href) {
			node.href = obj.Data.Href;
		}
		if (obj.Data.Target) {
			node.target = obj.Data.Target;
		}
		if (obj.Data.OnClick && OnClick[obj.Data.OnClick]) {
			OnClick[obj.Data.OnClick](node);
		}
		if (obj.Data.Focus === "true") {
			elem.focus();
		}
   		elem.appendChild(node);
		if (obj.Data.Scroll && obj.Data.Scroll == "true") {
			elem.scrollTop = elem.scrollHeight;
		}
	}
}
DomMap["innerHTML"] = function (elem, obj) {
	if (obj.Data.Value) {
		elem.innerHTML = obj.Data.Value;
	}
}
DomMap["editable"] = function (elem, obj) {
	if (obj.Data.Value) {
		if (obj.Data.Value === "true") {
			elem.contentEditable = true;
		} else {
			elem.contentEditable = false;
		}
	}
}
DomMap["focus"] = function (elem, obj) {
	if (obj.Data.Value) {
		if (obj.Data.Value === "true") {
			elem.focus();
		} else {
			elem.blur();
		}
	}
}
DomMap["setAttribute"] = function (elem, obj) {
	if (obj.Data.Attribute && obj.Data.Value) {
		elem.setAttribute(obj.Data.Attribute, obj.Data.Value);
	}
}
DomMap["getAttribute"] = function (elem, obj) {
	if (obj.Data.Attribute) {
		ws.send(elem.getAttribute(obj.Data.Attribute));
	}
}
DomMap["getProperty"] = function (elem, obj) {
	if (obj.Data.Property) {
		ws.send(window.getComputedStyle(elem,null).getPropertyValue(obj.Data.Property));
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
	if (obj.Data.Value) {
		elem.style.background = obj.Data.Value;
	}
}
DomMap["background-color"] = function (elem, obj) {
	if (obj.Data.Value) {
		elem.style.backgroundColor = obj.Data.Value;
	}
}
DomMap["color"] = function (elem, obj) {
	if (obj.Data.Value) {
		elem.style.color = obj.Data.Value;
	}
}
DomMap["border"] = function (elem, obj) {
	if (obj.Data.Value) {
		elem.style.border = obj.Data.Value;
	}
}
DomMap["border-color"] = function (elem, obj) {
	if (obj.Data.Value) {
		elem.style.borderColor = obj.Data.Value;
	}
}