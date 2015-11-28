package webserver

import (
	"../server"
	"log"
)

var fm *FeedsManager

var gId = &server.GroupId{}

type FeedsManager struct {
	join     chan *connection
	leave    chan *connection
	incoming chan *server.FeedItem
	clients  map[server.GroupId]map[*connection]bool
}

func NewFeedsManager() *FeedsManager {
	fm := &FeedsManager{
		join:     make(chan *connection),
		leave:    make(chan *connection),
		incoming: make(chan *server.FeedItem),
		clients:  make(map[server.GroupId]map[*connection]bool),
	}
	fm.Listen()
	return fm
}

func (fm *FeedsManager) Listen() {
	go func() {
		for {
			select {
			case client := <-fm.join:
				fm.Join(client)
			case client := <-fm.leave:
				fm.Leave(client)
			case message := <-fm.incoming:
				if err := server.HandleFeedItem(message); err == nil {
					fm.Broadcast(message)
				} else {
					log.Println("could not handle message", message,
						",\ndue to error:", err.Error())
				}
			}
		}
	}()
}

func (fm *FeedsManager) Join(client *connection) {
	if clientsThisFeed, exists := fm.clients[client.gId]; exists {
		clientsThisFeed[client] = true
	} else {
		fm.clients[client.gId] = make(map[*connection]bool)
		fm.clients[client.gId][client] = true
	}

	log.Println("client joined")
}

func (fm *FeedsManager) Leave(client *connection) {
	groupKey := client.gId
	if clientsThisFeed, exists := fm.clients[groupKey]; exists {
		if _, ok := clientsThisFeed[client]; ok {
			delete(clientsThisFeed, client)
			close(client.outgoing)
			log.Println("client left")

			if len(clientsThisFeed) == 0 {
				delete(fm.clients, groupKey)
				log.Println("unregistering feed for group ", groupKey)
			}
		}
	}
}

func (fm *FeedsManager) Broadcast(message *server.FeedItem) {

	for client := range fm.clients[message.GId] {
		client.outgoing <- message
	}
	log.Println("broadcasted message")
}
