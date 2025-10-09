package main

import (
	"bufio"
	. "fmt"
	// "gogo/engine"
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
				Println(session.Values["PlayerId"])
			}
			Store.Mu.RUnlock()
		default:
			Println("Comando no reconocido:", cmd)
		}
	}
}

func main() {

	// Game engine
	// ui.Draw()
	// engine.Play()

	// Web server

	InitDB()
	defer DB.Close()
	LoadTemplates()
	go cmd()
	go CleanupSessions()

	MainServer = NewServer() // for web sockets

	go MainServer.SendOnlinePlayers()
	go MainServer.PingLoop()

	mux := http.NewServeMux()
	mux.Handle("GET /login/{$}", GuestMiddleware(Template("login.html", nil)))
	mux.Handle("GET /test/{$}", GuestMiddleware(Template("test.html", nil)))
	mux.Handle("GET /home/{$}", LoggedMiddleware(Template("home.html", nil)))
	mux.Handle("GET /game/{id}", LoggedMiddleware(Template("game.html", nil)))
	mux.HandleFunc("POST /api/login/{$}", Login)
	mux.HandleFunc("POST /api/play/{$}", PlayWith)
	// mux.HandleFunc("GET /api/getallonlineplayers/{$}", GetAllOnlinePlayers)

	mux.Handle("/ws", websocket.Handler(MainServer.HandleWS))

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	err := http.ListenAndServe(":8080", SessionMiddleware(mux))
	if err != nil {
		Println("Error en servidor:", err)
		os.Exit(1)
	}
}
