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

type Page struct {
	Title string
	Body  []byte
}

var templates = template.Must(template.ParseFiles(
	"tmpl/htmx-index.html", "tmpl/edit.html", "tmpl/view.html"))

// Hash keys should be at least 32 bytes long
var hashKey = []byte("very-secret")

// Block keys should be 16 bytes (AES-128) or 32 bytes (AES-256) long.
// Shorter keys may weaken the encryption used.
var blockKey []byte = nil

var s = securecookie.New(hashKey, blockKey)

const cookieName = "go-htmx-playground"

func SetCookie(w http.ResponseWriter, r *http.Request) string {
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

func ReadCookie(w http.ResponseWriter, r *http.Request) (*string, error) {
	var err error
	if cookie, err := r.Cookie(cookieName); err == nil {
		value := make(map[string]string)
		if err = s.Decode(cookieName, cookie.Value, &value); err == nil {
			sessionUUID := value["uuid"]
			fmt.Printf("The value of sessionUUID is %q\n", sessionUUID)
			return &sessionUUID, nil
		}
	}
	return nil, err
}

func rootHandler(w http.ResponseWriter, r *http.Request, sessionStore sessions.LocalSessionStore) {
	var sessionUUID string
	mSessionUUID, _ := ReadCookie(w, r)
	if mSessionUUID == nil {
		sessionUUID = SetCookie(w, r)
	} else {
		sessionUUID = *mSessionUUID

	}
	userSession, err := sessionStore.Load(sessionUUID)
	if err != nil {
		fmt.Println("Unable to read session data for "+sessionUUID, err)
		userSession = &sessions.UserSession{
			Id:   sessionUUID,
			Data: 0,
		}
		fmt.Println("Initialized session data for " + sessionUUID)
		sessionStore.Save(*userSession)
	} else {
		fmt.Println("Found existing session data for ", (*userSession).Id)
	}
	err = templates.ExecuteTemplate(w, "htmx-index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func clickedHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<div>Clicked</div>")
	println("Clicked")
}

func main() {
	var localSessionStore = sessions.LocalSessionStore{
		Path: "data/sessions/",
	}

	rootHandlerSession := func(w http.ResponseWriter, r *http.Request) {
		rootHandler(w, r, localSessionStore)
	}

	http.HandleFunc("/", rootHandlerSession)
	http.HandleFunc("/clicked", clickedHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
