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
	// "sync"
)

func cmd() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmd := strings.TrimSpace(scanner.Text())
		switch cmd {
		case "active users":
			Store.Mu.RLock()
			for _, session := range Store.Sessions {
				Println(session.Values["UserId"])
			}
			Store.Mu.RUnlock()
		case "usernames":
			usernames := GetAllUsersUsernames()
			for _, username := range usernames {
				Println(username)
			}
		case "games":
			games := GetAllGames()
			for _, game := range games {
				user1 := GetUserById(game.UserId1)
				user2 := GetUserById(game.UserId2)
				Println(game.UserId1)
				Println(game.UserId2)
				Printf("Game id: %d || %s VS %s \n", game.Id, user1.Name, user2.Name)
			}
		default:
			if strings.HasPrefix(cmd, "user") {
				username := strings.TrimSpace(cmd[len("user"):])
				user := GetUserByUsername(username)
				if user == nil {
					Println("user does not exists")
				} else {
					Println(user)
				}
			} else {
				Println("Comando no reconocido:", cmd)
			}
		}
	}
}

func main() {

	// Game engine
	// ui.Draw()
	// engine.Play()

	// Web server

	err := InitDB()
	if err != nil {
		Println(err)
		return
	}

	defer DB.Close()
	LoadTemplates()
	go CleanupSessions()
	go cmd()


	MainServer = NewServer() // for web sockets

	go MainServer.SendOnlineUsers()
	go MainServer.PingLoop()

	mux := http.NewServeMux()
	mux.Handle("GET /login/{$}", GuestMiddleware(Template("login.html", nil)))
	mux.Handle("GET /test/{$}", GuestMiddleware(Template("test.html", nil)))
	mux.Handle("GET /home/{$}", LoggedMiddleware(Template("home.html", nil)))
	mux.Handle("GET /game/{id}", LoggedMiddleware(Template("game.html", nil)))
	mux.HandleFunc("POST /api/login/{$}", Login)
	mux.HandleFunc("POST /api/play/{$}", PlayWith)
	// mux.HandleFunc("GET /api/getallonlineusers/{$}", GetAllOnlineUsers)

	mux.Handle("/ws", websocket.Handler(MainServer.HandleWS))

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	err = http.ListenAndServe(":8080", SessionMiddleware(mux))
	if err != nil {
		Println("Error en servidor:", err)
		os.Exit(1)
	}
}
