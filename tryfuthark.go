package main

import (
	"io/ioutil"
	"net/http"
	"html/template"
	"regexp"
	"crypto/md5"
	"fmt"
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
var progPath = regexp.MustCompile("^/prog/([a-zA-Z0-9]+)$")

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
	p := make(map[string]string)
	p["Hash"] = hash
	p["Code"] = code
	renderTemplate(w, "prog", p)
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
	http.Redirect(w, r, "/prog/"+title, http.StatusFound)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "add", make(map[string]string))
}

func main() {
	http.HandleFunc("/prog/", progHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/add/", addHandler)
	http.ListenAndServe(":8080", nil)
}
