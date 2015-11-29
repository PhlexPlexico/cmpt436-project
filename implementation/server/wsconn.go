/*
Copyright (c) 2013 The Gorilla WebSocket Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

  Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

  Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

/*
Note: William included the above license, because this code is a copy of

https://github.com/gorilla/websocket/blob/master/examples/chat/conn.go

, with modifications.
*/
package server

import (
	"../db"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

const (
	// Time allowed to write a message to the client.
	writeWait = 10 * time.Second

	// 	// Time allowed to read the next pong message from the client.
	// 	pongWait = 60 * time.Second

	// 	// Send pings to client with this period. Must be less than pongWait.
	// 	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed.
	maxMessageSize = 512

	messageTypeFeedItem = "feedItem"
	messageTypeGroups   = "groups"
	messageTypeContacts = "contacts"
	messageTypeGroupId  = "groupId"
)

// connection is an middleman between the websocket connection and
// the Feeds Manager.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	//the user's unique Id.
	userId string

	// Buffered channel of outbound messages.
	outgoing chan *websocketOutMessage
}

type websocketOutMessage struct {
	Content json.RawMessage `json:"content"`
	Type    string          `json:"type"`
}

// readPump pumps messages from the websocket connection to the Feeds Manager.
func (c *connection) readPump() {
	defer func() {
		fm.leave <- c
		c.ws.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)

	/* We may not need to use pings and pongs. */
	// c.ws.SetReadDeadline(time.Now().Add(pongWait))
	// c.ws.SetPongHandler(func(string) error {
	// 	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	// 	return nil
	// })

	for {
		message := &db.FeedItem{}
		err := c.ws.ReadJSON(message)
		if err != nil {
			log.Println("error reading ws message: ", err)
			break
		}
		// if (message.GroupId == "") == (message.ContactsId == "") {
		// 	log.Println("bad message has both groupId and contactsId")
		// 	break
		// }

		fm.incoming <- message
	}
}

type writeFunc func() error

// writes a message with the given message type and payload.
func (c *connection) writeMessage(mt int, payload []byte) error {
	// c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.writeGeneric(func() error { return c.ws.WriteMessage(mt, payload) })
}

func (c *connection) writeWebsocketMessage(msg *websocketOutMessage) error {
	return c.writeGeneric(func() error { return c.ws.WriteJSON(msg) })
}

func (c *connection) writeGeneric(wf writeFunc) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return wf()
}

// writePump pumps messages from the Feeds Manager to the websocket connection.
func (c *connection) writePump() {
	// ticker := time.NewTicker(pingPeriod)
	defer func() {
		// ticker.Stop()
		log.Println("closing ws writer")
		c.ws.Close()
	}()
	for {
		// select {
		// case message, ok := <-c.outgoing:
		message, ok := <-c.outgoing
		log.Println("received outgoing message")
		if !ok {
			// c.writeMessage(websocket.CloseMessage, []byte{})
			return
		}
		if err := c.writeWebsocketMessage(message); err != nil {
			log.Println("error writing ws message: ", err)
			return
		}
		// case <-ticker.C:
		// 	if err := c.writeMessage(websocket.PingMessage, []byte{}); err != nil {
		// 		log.Println("error sending ping: ", err)
		// 		return
		// 	}
		// }
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(ws *websocket.Conn, userId string) {
	c := &connection{
		outgoing: make(chan *websocketOutMessage, 256),
		userId:   userId,
		ws:       ws,
	}

	fm.join <- c
	go c.writePump()
	c.readPump()
}
