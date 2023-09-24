package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/morucci/go-htmx/sessions"

	"github.com/gorilla/securecookie"
)

type Page struct {
	Title string
	Body  []byte
}

var templates = template.Must(template.ParseFiles(
	"tmpl/htmx-index.html", "tmpl/edit.html", "tmpl/view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// Hash keys should be at least 32 bytes long
var hashKey = []byte("very-secret")

// Block keys should be 16 bytes (AES-128) or 32 bytes (AES-256) long.
// Shorter keys may weaken the encryption used.
var blockKey []byte = nil

var s = securecookie.New(hashKey, blockKey)

func SetCookie(w http.ResponseWriter, r *http.Request) {
	println("Here set cookie")
	value := map[string]string{
		"foo": "bar",
	}
	if encoded, err := s.Encode("cookie-name", value); err == nil {
		cookie := http.Cookie{
			Name:     "cookie-name",
			Value:    encoded,
			Path:     "",
			Secure:   true,
			HttpOnly: true,
			MaxAge:   3600,
			SameSite: http.SameSiteLaxMode,
		}
		fmt.Println(cookie)
		http.SetCookie(w, &cookie)
	}
}

func ReadCookie(w http.ResponseWriter, r *http.Request) {
	println("Here read cookie")
	if cookie, err := r.Cookie("cookie-name"); err == nil {
		value := make(map[string]string)
		if err = s.Decode("cookie-name", cookie.Value, &value); err == nil {
			fmt.Printf("The value of foo is %q\n", value["foo"])
		}
	}
}

func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{
		Title: title,
		Body:  body,
	}, nil
}
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/index", http.StatusFound)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func indexHTMXHandler(w http.ResponseWriter, r *http.Request) {
	ReadCookie(w, r)
	SetCookie(w, r)
	err := templates.ExecuteTemplate(w, "htmx-index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func clickedHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<div>Clicked</div>")
	println("Clicked")
}

func main() {
	fmt.Println("Hello, World")
	userSession := sessions.UserSession{
		Id: "123",
	}
	localSessionStore := sessions.LocalSessionStore{
		Path: "data/sessions/",
	}

	err := localSessionStore.Save(userSession)
	if err != nil {
		fmt.Println("Unable to write session", err)
	}
	userSession2, err := localSessionStore.Load(userSession.Id)
	if err != nil {
		fmt.Println("Unable to read session", err)
	} else {
		fmt.Println("UserSession ID", (*userSession2).Id)
	}
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/htmx", indexHTMXHandler)
	http.HandleFunc("/clicked", clickedHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
