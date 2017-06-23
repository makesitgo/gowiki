package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
)

// templates pre-loads all html templates at startup
// this will panic if an error occurs and will exit the program
var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))

// validPath sets regular expression matcher for valid endpoints of our program
// this is to prevent any file being able to be read/written to our server
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// rootHandler redirects root path to /view/FrontPage
func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

// viewHandler loads wiki page and renders it in browser
// via the url pattern: /view/{Page.Title}
// if the page does not exist, request redirects to edit new Page
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

// editHandler provides form to edit and save wiki Page contents
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

// saveHandler saves Page to disk and redirects to view Page
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

// makeHandler consolidates the URL parsing logic to grab Page title
// and then executes fn with title paramter included
// if title is invalid or not found, an HTTP Not Found error is returned
func makeHandler(fn func(w http.ResponseWriter, r *http.Request, title string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

// renderTemplate consolidates processing involved with template rendering
// by executing provided Page on '{{tmpl}}.html' and writing to http response
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Page represents a standard, interconnected wiki page
// consisting of a title and body (the page content)
type Page struct {
	Title string
	Body  []byte
}

// save creates/updates a .txt file, named after this Page's Title
// and puts its Body as the file contents
func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

// loadPage constructs a .txt file name from the provided title,
// and loads the contents of that file (along with the title) into a Page
func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.ListenAndServe(":8080", nil)
}
