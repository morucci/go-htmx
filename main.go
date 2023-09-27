package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/morucci/go-htmx/sessions"

	"github.com/gorilla/securecookie"
)

var templates = template.Must(template.ParseFiles(
	"tmpl/htmx-index.html"))

// Hash keys should be at least 32 bytes long
var hashKey = []byte("very-secret")

// Block keys should be 16 bytes (AES-128) or 32 bytes (AES-256) long.
// Shorter keys may weaken the encryption used.
var blockKey []byte = nil

var s = securecookie.New(hashKey, blockKey)

func setCookie(w http.ResponseWriter, r *http.Request, cookieName string) string {
	sessionUUID := uuid.NewString()
	value := map[string]string{
		"uuid": sessionUUID,
	}
	encoded, _ := s.Encode(cookieName, value)
	cookie := http.Cookie{
		Name:     cookieName,
		Value:    encoded,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		MaxAge:   3600,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	return sessionUUID
}

func readCookieSessionUUID(w http.ResponseWriter, r *http.Request, cookieName string) string {
	if cookie, err := r.Cookie(cookieName); err == nil {
		value := make(map[string]string)
		if err = s.Decode(cookieName, cookie.Value, &value); err == nil {
			return value["uuid"]
		}
	}
	return ""
}

func (s *SessionHandler) getUserSession(w http.ResponseWriter, r *http.Request) sessions.UserSession {
	const cookieName = "go-htmx-playground"
	var sessionUUID string
	mSessionUUID := readCookieSessionUUID(w, r, cookieName)
	if mSessionUUID == "" {
		sessionUUID = setCookie(w, r, cookieName)
	} else {
		sessionUUID = mSessionUUID
	}
	userSession, err := s.sessionsStore.Load(sessionUUID)
	if err != nil {
		fmt.Println("Unable to read session data for "+sessionUUID, err)
		userSession = &sessions.UserSession{
			Id: sessionUUID,
			Data: sessions.UserData{
				Counter: 0,
			},
		}
		fmt.Println("Initialized session data for " + sessionUUID)
		s.sessionsStore.Save(*userSession)
	} else {
		fmt.Println("Found existing session data for ", (*userSession).Id)
	}
	return *userSession
}

func (s *SessionHandler) rootHandler(w http.ResponseWriter, r *http.Request) {
	userSession := s.getUserSession(w, r)
	err := templates.ExecuteTemplate(w, "htmx-index.html", userSession.Data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *SessionHandler) clickedButtonHandler(w http.ResponseWriter, r *http.Request, val int) {
	userSession := s.getUserSession(w, r)
	userSession.Data.Counter += val
	s.sessionsStore.Save(userSession)
	fmt.Fprintf(w, "<div id=counter-value>%d</div>", userSession.Data.Counter)
}

func (s *SessionHandler) clickedPlusButtonHandler(w http.ResponseWriter, r *http.Request) {
	s.clickedButtonHandler(w, r, +1)
}

func (s *SessionHandler) clickedMinusButtonHandler(w http.ResponseWriter, r *http.Request) {
	s.clickedButtonHandler(w, r, -1)
}

type SessionHandler struct {
	sessionsStore sessions.LocalSessionStore
}

func main() {

	s := SessionHandler{
		sessionsStore: sessions.LocalSessionStore{
			Path: "data/sessions/",
		},
	}

	http.HandleFunc("/", s.rootHandler)
	http.HandleFunc("/clicked-plus", s.clickedPlusButtonHandler)
	http.HandleFunc("/clicked-minus", s.clickedMinusButtonHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
