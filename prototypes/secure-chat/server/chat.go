package main

import (
	"github.com/garyburd/redigo/redis"
	"golang.org/x/net/websocket"
	"log"
	"strconv"
)

var (
	chat *Chat
)

type Client struct {
	ws *websocket.Conn
}

func NewClient(ws *websocket.Conn) *Client {
	client := &Client{
		ws: ws,
	}
	return client
}

type Message struct {
	Username string `json:"username" redis:"username"`
	Time     string `json:"time" redis:"time"`
	Message  string `json:"message" redis:"message"`
}

type Chat struct {
	join     chan *Client
	leave    chan *Client
	incoming chan *Message
	outgoing chan *Message
	clients  []*Client
	rc       redis.Conn
}

func NewChat() *Chat {
	chat := &Chat{
		join:     make(chan *Client),
		leave:    make(chan *Client),
		incoming: make(chan *Message),
		outgoing: make(chan *Message),
		clients:  make([]*Client, 0),
		rc:       pool.Get(),
	}
	chat.Listen()
	return chat
}

func (chat *Chat) Listen() {
	go chat.Subscribe()
	go func() {
		for {
			select {
			case client := <-chat.join:
				chat.Join(client)
			case client := <-chat.leave:
				chat.Leave(client)
			case message := <-chat.incoming:
				chat.Publish(message)
			case message := <-chat.outgoing:
				chat.Broadcast(message)
			}
		}
	}()
}

func (chat *Chat) Subscribe() {
	psc := redis.PubSubConn{pool.Get()}
	psc.Subscribe("chat")
	c := pool.Get()
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			log.Printf("%s: message %s\n", v.Channel, v.Data)
			id, err := redis.Int(v.Data, nil)
			if err != nil {
				log.Println(err)
				return
			}
			result, err := redis.Values(c.Do("HGETALL", "message:"+strconv.Itoa(id)))
			if err != nil {
				log.Println(err)
				return
			}

			var message Message
			err = redis.ScanStruct(result, &message)
			if err != nil {
				log.Println(err)
				return
			}
			chat.outgoing <- &message

		case redis.Subscription:
			log.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			log.Println(v)
			return
		}
	}
}

func (chat *Chat) Publish(message *Message) {
	id, err := redis.Int(chat.rc.Do("INCR", "message"))
	if err != nil {
		log.Println(err)
		return
	}
	_, _ = chat.rc.Do("HMSET", "message:"+strconv.Itoa(id),
		"username", message.Username,
		"time", message.Time,
		"message", message.Message)
	_, _ = chat.rc.Do("LPUSH", "messages", id)
	_, _ = chat.rc.Do("PUBLISH", "chat", id)
}

func (chat *Chat) Join(client *Client) {
	chat.clients = append(chat.clients, client)
	log.Println("client joined")
}

func (chat *Chat) Leave(client *Client) {
	chat.clients = append(chat.clients, client)
	for i, otherClient := range chat.clients {
		if client == otherClient {
			chat.clients = append(chat.clients[:i], chat.clients[i+1:]...)
			break
		}
	}
	log.Println("client left")
}

func (chat *Chat) Broadcast(message *Message) {
	for _, client := range chat.clients {
		websocket.JSON.Send(client.ws, message)
	}
	log.Println("broadcasted message")
}
