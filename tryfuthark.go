package main

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
)

// This function feels a little superfluous.
func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var templates = template.Must(template.ParseFiles("prog.html", "add.html"))

// Need some initialisation code to ensure this directory always
// exists.
var progdir = "prog"
var progPath = regexp.MustCompile("^/([a-zA-Z0-9]+)$")

func loadProg(hash string) (string, error) {
	filename := progdir + "/" + hash + ".fut"
	prog, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(prog[:]), nil
}

func saveProg(hash string, code string) error {
	filename := progdir + "/" + hash + ".fut"
	return ioutil.WriteFile(filename, []byte(code), 0600)
}

func progHandler(w http.ResponseWriter, r *http.Request) {
	m := progPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return
	}
	hash := m[1]
	code, err := loadProg(hash)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "prog", struct {
		Hash string
		Code string
	}{
		hash,
		code,
	})
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("program")
	title := fmt.Sprintf("%x", md5.Sum([]byte(body[:])))
	err := saveProg(title, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Here we should also spawn some process to compile the
	// program.
	http.Redirect(w, r, "/"+title, http.StatusFound)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	// Providing an empty interface{} is quicker than making
	// an empty map[string]string.
	renderTemplate(w, "add", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// We are trying to save a program
		saveHandler(w, r)
		return
	}
	if r.URL.Path != "/" {
		// We are looking for some program
		progHandler(w, r)
		return
	}
	// If nothing else, then we are just trying to add one.
	addHandler(w, r)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

