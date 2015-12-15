# Soshell
Soshell is a web-based interactive console written in Go & JavaScript. The working goal is to create a public web-based social platform with a chatroom/command console inspired interface.

## Basic Features
* Uses HTTPS/WSS for secure web connections.
* Simple command system for interacting with the server.
* Embedded Go-based server-side database (Tiedot).
* JavaScript/HTML/CSS client frontend.

## Usage

### Server Command Flags
	-cert 	- (required)            Path to encryption certificate.
	-key 	- (required)            Path to encryption key.
	-http 	- (default:80)          Web http port.
	-https 	- (default:443)         Web https port.
	-host	- (default:"localhost") Domain or host name.
	-public - (default:"public")    Public web directory path.
	-dbpath - (default:"database")  Path to database.
	-help	- Show command help information.

### Example
```
soshell -host="example.com" -http=8080 -https=8090 -cert="/dir/ssl/example.com/fullchaim.pem" -key="/dir/ssl/example.com/privkey.pem" -dbpath="/dir/db"
```
