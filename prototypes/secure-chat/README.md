# secure-chat

This is intended to be a simple prototype for testing https and wss, using a Redis backend.

### Setup

Setup is a bit more involved for this prototype. Make sure you have [Bower](http://bower.io/), [Go](https://golang.org/), [Redis](http://redis.io/), and [OpenSSL](https://www.openssl.org/) installed (preferably using [Homebrew](http://brew.sh/) or apt-get if possible), and then run the following commands:

1. `cd app`
2. `bower install`
	* Installs required web components
3. `cd ../server`
4. `sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout key.key -out cert.crt`
 * Generates a self-signed certificate for https and wss
5. `redis-server`
 * Starts a redis server, **do this in a seperate window**
6. `go run *.go`

You may need to install some Go packages, in which case you can usually type in `go get foo`, where foo is the name of the missing package that needs to be installed. 

After all this is done, the site will be accessible via [http://localhost:8000/](http://localhost:8000/)

### Notes

* The official Golang websocket implementation, [golang.org/x/net/websocket](https://godoc.org/golang.org/x/net/websocket) does not contain any way of limiting the size of the messages sent over the network
  * It may be prefferable to use [github.com/gorilla/websocket](https://godoc.org/github.com/gorilla/websocket) since it supports this feature which is important so that users can't just send giant strigns and crash our server
