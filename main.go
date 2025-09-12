package main

import (
	"bufio"
	. "fmt"
	. "go-online/globals"
	. "go-online/handlers"
	. "go-online/utils"
	. "go-online/engine"
	"golang.org/x/net/websocket"
	"net/http"
	"os"
	"strings"
	// "time"
	// bolt "go.etcd.io/bbolt"
)

func cmd() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmd := strings.TrimSpace(scanner.Text())
		switch cmd {
		case "users":
			Store.Mu.RLock()
			for _, session := range Store.Sessions {
				Println(session.Values["name"])
			}
			Store.Mu.RUnlock()
		default:
			Println("Comando no reconocido:", cmd)
		}
	}
}

func main() {

	// Game engine
	Play()
	return

	// Web server

	InitDB()
	defer DB.Close()
	LoadTemplates()
	go cmd()
	go CleanupSessions()

	server := NewServer() // for web sockets
	go server.SendOnlinePlayers()
	go server.PingLoop()

	mux := http.NewServeMux()
	mux.Handle("GET /login/{$}", GuestMiddleware(Template("login.html", nil)))
	mux.Handle("GET /test/{$}", GuestMiddleware(Template("test.html", nil)))
	mux.Handle("GET /home/{$}", LoggedMiddleware(Template("home.html", nil)))
	mux.HandleFunc("POST /api/login/{$}", Login)
	// mux.HandleFunc("GET /api/getallonlineplayers/{$}", GetAllOnlinePlayers)

	mux.Handle("/ws", websocket.Handler(server.HandleWS))

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	err := http.ListenAndServe(":8080", SessionMiddleware(mux))
	if err != nil {
		Println("Error en servidor:", err)
		os.Exit(1)
	}
}
