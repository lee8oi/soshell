# Soshell
Soshell is a web-based interactive console written in Go & JavaScript. The working goal is to create a public web-based social platform with a chatroom/command console inspired interface.

## Basic Features
* Uses HTTPS/WSS for secure web connections.
* Simple command system for interacting with the server.
* Embedded Go-based server-side database (Tiedot).
* JavaScript/HTML/CSS client frontend.

## Usage

### Server Command Flags
-http 	- Web http port.
-https 	- Web https port.
-host	- Domain or host name.
-public - Public web directory path.
-cert 	- Path to encryption certificate.
-key 	- Path to encryption key.
-dbpath - Path to database.
-help	- Show command help information.

### Example
```
soshell -host="example.com" -http=8080 -https=8090 -cert="/dir/ssl/example.com/fullchaim.pem" -key="/dir/ssl/example.com/privkey.pem" -dbpath="/dir/db"
```