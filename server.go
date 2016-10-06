package main

import (
    "github.com/codegangsta/martini"
    "github.com/martini-contrib/render"
    "github.com/gorilla/websocket"
    "net/http"
    "log"
    "net"
    "sync"
)

var ActiveClients = make(map[ClientConn] int)
var ActiveClientsRWMutex sync.RWMutex

type ClientConn struct {
    websocket *websocket.Conn
    clientIP  net.Addr
}

func addClient(cc ClientConn) {
    ActiveClientsRWMutex.Lock()
    ActiveClients[cc] = 0
    ActiveClientsRWMutex.Unlock()
}

func deleteClient(cc ClientConn) {
    ActiveClientsRWMutex.Lock()
    delete(ActiveClients, cc)
    ActiveClientsRWMutex.Unlock()
}

func broadcastMessage(messageType int, message []byte) {
    ActiveClientsRWMutex.RLock()
    defer ActiveClientsRWMutex.RUnlock()

    for client, _ := range ActiveClients {
        err := client.websocket.WriteMessage(messageType, message)
        if err != nil {
            log.Println(err)
            return
        }
    }
}

func main() {
    m := martini.Classic()
    m.Use(render.Renderer())
    m.Get("/", func(r render.Render) {
        r.HTML(200, "index", nil)
    })
    m.Get("/sock", func(w http.ResponseWriter, r *http.Request) {
        log.Println(ActiveClients)
        ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
        // error handling
        if _, ok := err.(websocket.HandshakeError); ok {
            http.Error(w, "Not a websocket handshake", 400)
            return
        } else if err != nil {
            log.Println(err)
            return
        }
        log.Println("here")
        clientAddr := ws.RemoteAddr()
        sockClient := ClientConn{ws, clientAddr}
        addClient(sockClient)

        for {
            log.Println(len(ActiveClients), ActiveClients)
            messageType, p, err := ws.ReadMessage()
            if err != nil {
                deleteClient(sockClient)
                log.Println("bye")
                log.Println(err)
                return
            }
            broadcastMessage(messageType, p)
        }
    })
    m.Run()
}