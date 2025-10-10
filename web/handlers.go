package main

import (
	. "fmt"
	"net/http"
	// "slices"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"
	. "strconv"
)

func GetCurrentLang(r *http.Request) string {
	// lang, ok := r.Context().Value("lang").(string)
	// if ok && lang != "" {
	// 	return lang
	// }

	// cookie, err := r.Cookie("lang")

	// if err == nil && slices.Contains(LangsName, cookie.Value) {
	// 	return cookie.Value
	// }

	return "en"
}

func GenerateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func CleanupSessions() {
	for {
		time.Sleep(time.Hour)
		Store.Mu.Lock()
		for id, session := range Store.Sessions {
			if time.Now().After(session.ExpiresAt) {
				delete(Store.Sessions, id)
			}
		}
		Store.Mu.Unlock()
	}
}

func LoggedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		var session Session

		if err == nil {
			Store.Mu.RLock()
			stored, exists := Store.Sessions[cookie.Value]
			Store.Mu.RUnlock()

			if exists && stored.ExpiresAt.After(time.Now()) {
				session = stored
			} else {
				Store.Mu.Lock()
				delete(Store.Sessions, cookie.Value)
				Store.Mu.Unlock()
				http.SetCookie(w, &http.Cookie{
					Name:   "session_id",
					Value:  "",
					MaxAge: -1,
				})
				http.Redirect(w, r, "/login/", http.StatusFound)
				return
			}
		} else {
			http.Redirect(w, r, "/login/", http.StatusFound)
			return
		}

		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		var session Session

		if err == nil {
			Store.Mu.RLock()
			// extraemos session
			stored, exists := Store.Sessions[cookie.Value]
			Store.Mu.RUnlock()

			if exists {
				if stored.ExpiresAt.After(time.Now()) {
					// Si no ha expirado
					session = stored
				} else {
					// Se borra si expira
					Store.Mu.Lock()
					delete(Store.Sessions, cookie.Value)
					Store.Mu.Unlock()
					http.SetCookie(w, &http.Cookie{
						Name:   "session_id",
						Value:  "",
						MaxAge: -1, // El browser la elimina automaticamente
					})
				}
			}
		}

		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GuestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := r.Context().Value("session").(Session)
		if ok && session.ExpiresAt.After(time.Now()) {
			http.Redirect(w, r, "/home/", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req UserLoginReq
	if !ParseRequest(w, r, &req) {
		return
	}

	new_session_id, _ := GenerateSessionID()
	values := make(map[string]string)
	values["name"] = req.Name

	user := User{
		Name:        req.Name,
		InGame:      false,
		LastConnect: time.Now(),
	}

	id, _ := CreateUser(user)
	Printf("id of the user created: %d with ITOA: %d \n", id, Itoa(int(id)))
	values["UserId"] = Itoa(int(id))

	new_session := Session{
		ID:        new_session_id,
		Values:    values,
		ExpiresAt: time.Now().Add(60 * time.Minute),
	}

	Printf("New session: %#v \n", new_session)
	Store.Mu.Lock()
	Store.Sessions[new_session.ID] = new_session
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    new_session.ID,
		Path:     "/",
		Expires:  new_session.ExpiresAt,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	Store.Mu.Unlock()

	Redirect(w, r, "/home/")
	return
}


func StrToUint64(s string) uint64 {
	n, err := ParseUint(s, 10, 64)
	if err != nil {
		Println(err)
	}
	return n
}

func PlayWith(w http.ResponseWriter, r *http.Request) {
	var req PlayReq
	if !ParseRequest(w, r, &req) {
		return
	}


	cookie, err := r.Cookie("session_id")
	if err != nil {
		// TODO send to error
		return
	}


	user1 := StrToUint64(Store.Sessions[cookie.Value].Values["UserId"])
	user2 := req.Id

	//TODO check that the user has correct credentials session
	Printf("user1: %d \n", user1)
	Printf("user2: %d \n", user2)



	game := Game{
		UserId1: uint64(user1),
		UserId2: uint64(user2),
	}

	game_id, err := CreateGame(game)
	if err != nil {
		Println(err)
		return
	}
	Println(game_id)


	// j1 := map[string]string{"play": user2}
	// j2 := map[string]string{"play": user1}
	// msg1, _ := json.Marshal(j1)
	// msg2, _ := json.Marshal(j2)

	// MainServer.SendToUser(user1, string(msg1))
	// MainServer.SendToUser(user2, string(msg2))
	route := "game/" + Itoa(int(game_id))
	SendRedirectWS(user1, route)
	SendRedirectWS(user2, route)

	w.WriteHeader(http.StatusOK)
}


func SendRedirectWS(user_id uint64, route string) {
	j := map[string]string{"redirect": route}
	msg, _ := json.Marshal(j)
	MainServer.SendToUser(user_id, string(msg))
}


/*
- Build the template and think how to pass the data to the html about the current game state
*/