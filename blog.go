//package blog
package main

import (
	"os"
	"log"
	"html/template"
	"path"
	"net/http"
	"io/ioutil"
	)

type Blog struct {
	Title	string
	Body	[]byte
}

const (
	blogPrefix = "f"
	lenPath = len("/view/")
	)

func (b *Blog) save() error {
	filename := path.Join(blogPrefix, b.Title + ".txt")
	return ioutil.WriteFile(filename, b.Body, 0600)
}

func load(title string) (*Blog, error) {
	filename := path.Join(blogPrefix, title + ".txt")
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Blog{Title: title, Body: body}, nil
}

func index(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if _, ok := recover().(error); ok {
			log.Println("Failed to init blog")
		}
	}()
	d, err := os.Open(blogPrefix)
	if err != nil {
		panic(err)
	}

	files, err := d.Readdir(-1)
	if err != nil {
		panic(err)
	}

	list := make([]Blog, 0)
	for i := range files {
		if !files[i].IsDir() {
			name := files[i].Name()
			b := Blog{name[:len(name)-len(path.Ext(name))], nil}
			list = append(list, b)
		}
	}
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, list)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[lenPath:]
	p, _ := load(title)
	t, _ := template.ParseFiles("view.html")
	t.Execute(w, p)
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/view/", viewHandler)
	http.ListenAndServe(":8080", nil)
}
