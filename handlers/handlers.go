package handlers

import (
	. "fmt"
	// . "gogo/strings"
	"net/http"
	// "slices"
	"context"
	"crypto/rand"
	"encoding/base64"
	. "gogo/globals"
	. "gogo/models"
	// . "gogo/utils"
	"time"
)

func redirect(w http.ResponseWriter, route string) {
	w.Header().Set("HX-Redirect", route)
  w.WriteHeader(http.StatusOK)
}

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
			// No hay cookie, redirigir
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
	var req UserReq
	err := ParseForm(r, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Solicitud inválida: " + err.Error()))
		return
	}

	new_session_id, _ := GenerateSessionID()
	values := make(map[string]string)
	values["name"] = req.Name

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

	player := Player {
		Name: req.Name,
		InGame: false,
		LastConnect: time.Now(),
	}

	SavePlayer(player)

	// Using htmx
	redirect(w, "/home/")
	return
}

// func Home(w http.ResponseWriter, r *http.Request) {

// }

// func GetAllOnlinePlayers(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "text/html")

// 	Store.Mu.RLock()
// 	players := make([]string, 0, len(Store.Sessions))
// 	for _, session := range Store.Sessions {
// 		if name, ok := session.Values["name"].(string); ok {
// 			players = append(players, name)
// 		}
// 	}
// 	Store.Mu.RUnlock()

// 	for _, player := range players {
// 		RenderTemplate(w, "playerlistdiv", map[string]interface{}{
// 			"Name": player,
// 		})
// 	}
// }
