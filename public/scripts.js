/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

/* 
This file contains the websocket functions along with additional functions that are used by
the server to allow interactive access to the client html/css/js.
*/

var ws;
var disconnected = false;
function startSock() {
	ws = new WebSocket(sockUrl);
	ws.onopen = function (event) {
		append("#msg-list", '<div class="msg">Connected</div>');
		disconnected = false;
		document.getElementById("msg-txt").focus();
	};
	ws.onclose = function(){
		if (!disconnected) {
			append("#msg-list", '<div class="msg">Disconnected</div>');
			disconnected = true;
		}
		setTimeout(startSock, 3000);
	};
	ws.onmessage = function(event) {
		if (funcs[event.data.split("(")[0]]) {
			eval(event.data);
		}
	};
}
startSock();
var funcs = {}; //set of functions allowed in eval
function Send() {
	var elem = document.getElementById("msg-txt")
	ws.send(elem.value);
	elem.value = "";
	return false
}
funcs["append"] = true;
function append(selector, elem) {
	$(selector).append(elem);
}
funcs["innerHTML"] = true;
function innerHTML(selector, text) {
	$(selector).html(text);
}
funcs["editable"] = true;
function editable(selector, val) {
	var elem = document.querySelector(selector)
	elem.contentEditable = val;
}
funcs["focus"] = true;
function focus(selector, val) {
	if (val === true) {
		$(selector).focus();
	} else {
		$(selector).blur();
	}
}
funcs["setAttribute"] = true;
function setAttribute(selector, attr, val) {
	$(selector).attr(attr, val);
}
funcs["getAttribute"] = true;
function getAttribute(selector, attr) {
	console.log(attr);
	ws.send($(selector).attr(attr));
}
funcs["getProperty"] = true;
function getProperty(selector, prop) {
	var elem = document.querySelector(selector);
	ws.send($(selector).css(prop));
}
funcs["exists"] = true;
function exists(selector) {
	if ($(selector)) {
		ws.send("true");
	} else {
		ws.send("false");
	}
}
funcs["getHTML"] = true;
function getHTML(selector) {
	ws.send($(selector).html());
}
funcs["setProperty"] = true;
function setProperty(selector, prop, val) {
	$(selector).css(prop, val);
}
funcs["scroll"] = true;
function scroll(selector) {
	var elem = document.querySelector(selector);
	elem.scrollTop = elem.scrollHeight;
}