# redigo-chat

This is intended to be a simple prototype for testing a simple chat server using Redis and WebSocket. The Redis backend hasn't been implemented yet, so the Go server is basically just sending messages to all connected users, but I plan on using [Pub/Sub](http://redis.io/topics/pubsub) once I have time.

### Setup

Setup is pretty basic. Make sure you have [Bower](http://bower.io/) and [Go](https://golang.org/) installed, and then just run:

`bower install`

in the `app/` directory, and

`go run main.go`

in the `server/` directory.

After that, the site will be accessible via http://localhost:8000/