package main

import (
    "log"
    "net/http"
    "sync"

    "github.com/fsnotify/fsnotify"
    "github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var clientsMutex sync.Mutex
var upgrader = websocket.Upgrader{}

func main() {
    // Start file watcher
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

    go watchFiles(watcher)

    // Serve static files
    http.Handle("/", http.FileServer(http.Dir(".")))

    // WebSocket endpoint
    http.HandleFunc("/reload", wsHandler)

    // Watch current directory
    err = watcher.Add(".")
    if err != nil {
        log.Fatal(err)
    }

    port := ":8000"
    log.Println("Serving at http://localhost" + port)
    log.Fatal(http.ListenAndServe(port, nil))
}

func watchFiles(watcher *fsnotify.Watcher) {
    for {
        select {
        case event := <-watcher.Events:
            log.Println("File changed:", event.Name)
            broadcastReload()
        case err := <-watcher.Errors:
            log.Println("Watcher error:", err)
        }
    }
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        return
    }
    defer conn.Close()

    clientsMutex.Lock()
    clients[conn] = true
    clientsMutex.Unlock()

    // Keep connection open
    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            clientsMutex.Lock()
            delete(clients, conn)
            clientsMutex.Unlock()
            break
        }
    }
}

func broadcastReload() {
    clientsMutex.Lock()
    defer clientsMutex.Unlock()

    for client := range clients {
        err := client.WriteMessage(websocket.TextMessage, []byte("reload"))
        if err != nil {
            log.Println("WebSocket error:", err)
            client.Close()
            delete(clients, client)
        }
    }
}
