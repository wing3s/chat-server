package main

import (
	"github.com/codegangsta/martini"
	"github.com/gorilla/websocket"
	"github.com/martini-contrib/render"
	"log"
	"net"
	"net/http"
	"sync"
)

type ChatRoom struct {
	activeClients map[ClientConn]int
	activeClientsRWMutex sync.RWMutex
	queue chan string
}

// chat room init()
func (cr *ChatRoom) Init() {
	cr.activeClients = make(map[ClientConn]int)
}

type ClientConn struct {
	websocket *websocket.Conn
	clientIP  net.Addr
}

func (cr *ChatRoom) AddClient(cc ClientConn) {
	defer cr.activeClientsRWMutex.Unlock()

	cr.activeClientsRWMutex.Lock()
	cr.activeClients[cc] = 0
}


func (cr *ChatRoom) DeleteClient(cc ClientConn) {
	cr.activeClientsRWMutex.Lock()
	delete(cr.activeClients, cc)
	cr.activeClientsRWMutex.Unlock()
}


func (cr *ChatRoom) BroadcastMessage(messageType int, message []byte) {
	cr.activeClientsRWMutex.RLock()
	defer cr.activeClientsRWMutex.RUnlock()

	for client, _ := range cr.activeClients {
		err := client.websocket.WriteMessage(messageType, message)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(chatRoom.activeClients)
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	// error handling
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	clientAddr := ws.RemoteAddr()
	sockClient := ClientConn{ws, clientAddr}
	chatRoom.AddClient(sockClient)

	go func() {
		for {
			log.Println(len(chatRoom.activeClients), chatRoom.activeClients)
			messageType, p, err := ws.ReadMessage()
			if err != nil {
				chatRoom.DeleteClient(sockClient)
				log.Println(err)
				return
			}
			chatRoom.BroadcastMessage(messageType, p)
		}
	}()

}

var chatRoom ChatRoom

func main() {
	m := martini.Classic()
	m.Use(render.Renderer())
	m.Get("/", func(r render.Render) {
		r.HTML(200, "index", nil)
	})
	chatRoom.Init()
	m.Get("/sock", socketHandler)
	m.Run()
}
