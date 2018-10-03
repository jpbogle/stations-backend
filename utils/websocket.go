package utils

import (
    // "encoding/json"
    "bytes"
    "net/http"
    "stations/entities"
    "github.com/gorilla/websocket"
    "encoding/base64"
    // "stations/controllers"
    // "sync"
    // "time"
    // "fmt"
    "log"

)

type websocketManager struct {
    Websockets   map[string]*Websocket
    // lock         sync.Mutex               // protects websocket
}

type Websocket struct {
    id              string
    broadcast       chan entities.StationBroadcast
    currentSong     entities.Playing
    listeners       map[*websocket.Conn]bool
    admins          map[*websocket.Conn]bool
    upgrader        websocket.Upgrader
}

var globalWebsockets *websocketManager

func init() {
    globalWebsockets = CreateWebsocketManager()
}

func CreateWebsocketManager() *websocketManager {
    manager := &websocketManager {
        Websockets: make(map[string]*Websocket, 0),
    }
    return manager
}

// func CreateSocket() *Websocket {
//     var upgrader = websocket.Upgrader{
//         CheckOrigin: func(r *http.Request) bool {
//             return true
//         },
//     }
//     socket := &Websocket{
//         listeners: make(map[*websocket.Conn]bool),
//         broadcast: make(chan entities.StationBroadcast),
//         nowPlaying: make(chan entities.Playing),
//         upgrader: upgrader,
//     }
//     return socket
// }



//////////////////////////////////////////////////////////////////////////////
//
// Public websocketManager methods
//
//////////////////////////////////////////////////////////////////////////////\
///
// Creates a new websocket or gets existing one from websocketManager
func (wm *websocketManager) OpenWebsocket(ctx *Context, station *entities.Station) {
    // wm.lock.Lock()
    id := wm.getSocketId(ctx)
    var socket *Websocket
    socket, ok := wm.Websockets[id]

    if !ok {
        var upgrader = websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                return true
            },
        }
        socket = &Websocket{
            id: id,
            broadcast: make(chan entities.StationBroadcast),
            listeners: make(map[*websocket.Conn]bool),
            admins: make(map[*websocket.Conn]bool),
            upgrader: upgrader,
        }
        wm.Websockets[id] = socket
        // wm.lock.Unlock()
        //Listen from client loop
        go socket.send()
        // go socket.checkTime(ctx.Fields["username"], ctx.Fields["stationName"])
        socket.addAdmin(ctx, id, station)
    } else {
        // wm.lock.Unlock()
        socket.addListener(ctx, station)
    }
}

func (wm *websocketManager) CloseWebsocketCtx(ctx *Context) {
    // wm.lock.Lock()
    id := wm.getSocketId(ctx)

    if socket, ok := wm.Websockets[id]; ok {
        for client := range socket.listeners {
            client.Close()
            delete(socket.listeners, client)
        }
        for client := range socket.admins {
            client.Close()
            delete(socket.admins, client)
        }
        delete(wm.Websockets, id)
    }
    // wm.lock.Unlock()

}


func (wm *websocketManager) CloseWebsocket(id string) {
    // wm.lock.Lock()
    // id := wm.getSocketId(ctx)

    if socket, ok := wm.Websockets[id]; ok {
        for client := range socket.listeners {
            client.Close()
            delete(socket.listeners, client)
        }
        for client := range socket.admins {
            client.Close()
            delete(socket.admins, client)
        }
        delete(wm.Websockets, id)
    }
    // wm.lock.Unlock()

}

func (wm *websocketManager) Broadcast(ctx *Context, station *entities.Station, header string, message string) {
    id := wm.getSocketId(ctx)
    if socket, ok := wm.Websockets[id]; ok {
        response := entities.StationBroadcast{
            Station: station,
            Header: header,
            Message: message,
        }
        socket.broadcast <- response
    }
}


//////////////////////////////////////////////////////////////////////////////
//
// Pritave websocketManager methods
//
//////////////////////////////////////////////////////////////////////////////

// Gets a base64URL encoding for a websocket base on :username/:stationName
func (wm *websocketManager) getSocketId(ctx *Context) string {
    var buffer bytes.Buffer
    buffer.WriteString(ctx.Fields["username"])
    buffer.WriteString("/")
    buffer.WriteString(ctx.Fields["stationName"])
    return base64.URLEncoding.EncodeToString(buffer.Bytes())
}


//////////////////////////////////////////////////////////////////////////////
//
// Pritave Websocket methods
//
//////////////////////////////////////////////////////////////////////////////

// Adds a new host to a given websocket
func (socket *Websocket) addAdmin(ctx *Context, id string, station *entities.Station) {
    ws, err := socket.upgrader.Upgrade(ctx.Res, ctx.Req, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()
    socket.admins[ws] = true

    log.Printf(
        "WS Host /%s/%s %s%v\x1b[0m\n",
        ctx.Fields["username"],
        ctx.Fields["stationName"],
        "\x1b[32m",
        200,
    )
    broadcast := entities.StationBroadcast{
        Station: station,
        Player: &socket.currentSong,
        Admin: true,
        Header: "Welcome",
        Message: "Welcome to the station! You are now the admin",
    }
    ws.WriteJSON(broadcast)
    socket.receivePlaying(ws, id)
}

// Adds a new listener to a given websocket
func (socket *Websocket) addListener(ctx *Context, station *entities.Station) {
    ws, err := socket.upgrader.Upgrade(ctx.Res, ctx.Req, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()
    socket.listeners[ws] = true

    log.Printf(
        "WS /%s/%s %s%v\x1b[0m\n",
        ctx.Fields["username"],
        ctx.Fields["stationName"],
        "\x1b[32m",
        200,
    )
    // View sockets and listeners
    // for socket := range globalWebsockets.Websockets {
    //     log.Println(socket, globalWebsockets.Websockets[socket].listeners)
    // }
    broadcast := entities.StationBroadcast{
        Station: station,
        Player: &socket.currentSong,
        Admin: false,
        Header: "Welcome",
        Message: "Welcome to the station!",
    }
    ws.WriteJSON(broadcast)
    socket.receiveNothing(ws)
}

func (socket *Websocket) receiveNothing(ws *websocket.Conn) {
    for {
        var nowPlaying entities.Playing
        // Read in a new message as JSON and map it to a nowPlaying object
        err := ws.ReadJSON(&nowPlaying)
        if err != nil {
            ws.Close()
            log.Printf("Socket receiving error: %v", err)
            delete(socket.listeners, ws)
            break
        }
        //Do not send channel if listener not admin
    }
}

func (socket *Websocket) receivePlaying(ws *websocket.Conn, id string) {
    for {
        var nowPlaying entities.Playing
        // Read in a new message as JSON and map it to a nowPlaying object
        err := ws.ReadJSON(&nowPlaying)
        if err != nil {
            ws.Close()
            log.Printf("Socket receiving error: %v", err)
            delete(socket.admins, ws)
            if (len(socket.admins) == 0) {
                log.Println("No admins: closing websocket")
                globalWebsockets.CloseWebsocket(id)
            }
            break
        }
        socket.currentSong = nowPlaying
        broadcast := entities.StationBroadcast{
            Player: &nowPlaying,
        }
        socket.broadcast <-  broadcast
    }
}

// func (socket *Websocket) sendPlaying() {
//     for {
//         nowPlaying := <- socket.songBroadcast
//         // Send it out to every client that is currently connected
//         for client := range socket.listeners {
//             err := client.WriteJSON(nowPlaying)
//             if err != nil {
//                 log.Printf("Socket sending error: %v", err)
//                 client.Close()
//                 delete(socket.listeners, client)
//             }
//         }
//         for client := range socket.admins {
//             err := client.WriteJSON(nowPlaying)
//             if err != nil {
//                 log.Printf("Socket sending error: %v", err)
//                 client.Close()
//                 delete(socket.listeners, client)
//             }
//         }
//     }
// }

func (socket *Websocket) send() {
    for {
        // Grab the station from the broadcast channel
        station := <- socket.broadcast
        // Send it out to every client that is currently connected
        for client := range socket.listeners {
            err := client.WriteJSON(station)
            if err != nil {
                log.Printf("Socket sending error: %v", err)
                client.Close()
                delete(socket.listeners, client)
            }
        }
        for client := range socket.admins {
            station.Admin = true
            err := client.WriteJSON(station)
            if err != nil {
                log.Printf("Socket sending error: %v", err)
                client.Close()
                delete(socket.admins, client)
            }
        }
    }
}

// func (socket *Websocket) checkTime(username string, stationName string) {
//     for {
//         id := socket.id
//         globalWebsockets.lock.Lock()
//         if socket, ok := globalWebsockets.Websockets[id]; ok {
//             if socket.currentSong.Song.Duration > 0 && socket.currentSong.Playing {
//                 position := int64(socket.currentSong.Position)
//                 end := int64(socket.currentSong.Song.Duration)
//                 if position + ((time.Now().UTC().UnixNano() / 1e6) - socket.currentSong.Timestamp) >= end {
//                     station, err := controllers.PlayNext(username, stationName)
//                     if err != nil {
//                         return
//                     }
//                     message := fmt.Sprintf("Now Playing %s - %s", station.Playing.Song.Title, station.Playing.Song.Artist)
//                     broadcast := entities.StationBroadcast{
//                         Station: station,
//                         Message: message,
//                         Player: &station.Playing,
//                     }
//                     socket.currentSong = station.Playing
//                     socket.broadcast <- broadcast
//                 }
//             }
//         } else {
//             globalWebsockets.lock.Unlock()
//             break;
//         }
//         globalWebsockets.lock.Unlock()
//     }
// }
