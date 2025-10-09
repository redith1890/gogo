package main

import (
	"encoding/json"
	. "fmt"
	"golang.org/x/net/websocket"
	"io"
	"time"
)

type ConnFlag int

const (
	DEFAULT        ConnFlag = 0
	GET_LIST_USERS          = iota
)


func NewServer() *Server {
	return &Server{
		conns:   make(map[*websocket.Conn]*SessionInfo),
		Changed: make(chan struct{}, 1),
	}
}

func (s *Server) HandleWS(ws *websocket.Conn) {
	Println("new incoming connection from client: ", ws.RemoteAddr())
	id := GetPlayerIdFromRequest(ws)
	s.conns[ws] = &SessionInfo{
		Id:  id,
		Flags: DEFAULT,
	}
	s.MarkChanged()
	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)

	ws.SetReadDeadline(time.Now().Add(30 * time.Second))
	for {
		n, err := ws.Read(buf)
		if err != nil {
			delete(s.conns, ws)
			ws.Close()
			s.MarkChanged()
			Println("read error:", err)
			if err != io.EOF {
				Println("read error:", err)
			}
			break
		}
		msg := string(buf[:n])

		switch msg {
		case "GetAllOnlinePlayers":
			s.conns[ws].Flags |= GET_LIST_USERS
			s.MarkChanged()
		// Not used
		case "pong":
			Println("pong recibido")
			ws.SetReadDeadline(time.Now().Add(30 * time.Second))
		}

	}
	ws.Close()
	delete(s.conns, ws)
	s.MarkChanged()
}

// Not used
func (s *Server) PingLoop() {
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		Println(s.conns)
		for conn := range s.conns {
			if err := websocket.Message.Send(conn, "ping"); err != nil {
				Println("Error enviando ping, cerrando conexión:", err)
				conn.Close()
				delete(s.conns, conn)
				s.MarkChanged()
			}
		}
	}
}

func (s *Server) MarkChanged() {
	select {
	case s.Changed <- struct{}{}:
	default:
	}
}

func GetPlayerIdFromRequest(ws *websocket.Conn) uint64 {
	req := ws.Request()

	cookie, err := req.Cookie("session_id")
	if err != nil {
		return 0
	}

	sessionID := cookie.Value

	Store.Mu.RLock()
	defer Store.Mu.RUnlock()

	if session, ok := Store.Sessions[sessionID]; ok {
		return StrToUint64(session.Values["PlayerId"])
	}

	return 0
}

func (s *Server) SendOnlinePlayers() {
	for range s.Changed {
		var players []uint64
		for _, info := range s.conns {
			players = append(players, info.Id)
		}
		data, _ := json.Marshal(players)
		for conn, info := range s.conns {
			if info.Flags&GET_LIST_USERS != 0 {
				conn.Write(data)
			}
		}
	}
}

func (s *Server) SendToPlayer(PlayerId uint64, message string) error {
    sChanged := false
    for conn, info := range s.conns {
        if info.Id == PlayerId {
            err := websocket.Message.Send(conn, message)
            if err != nil {
                return err
            }
            sChanged = true
        }
    }
    if !sChanged {
        return Errorf("player not connected")
    }
    return nil
}
