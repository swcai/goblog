package main

import (
//	"os"
	"log"
	"html/template"
//	"path"
	"net/http"
//	"io/ioutil"
	"encoding/json"
	"time"
//	"sort"
	"strconv"
//	"bytes"
	"github.com/simonz05/godis"
)


type BlogEntry struct {
	Title	string
	Body	[]byte
	Date	int64
}

type BlogStruct struct {
	Name	string
	Author	string
	Email	string
	Entries	[]BlogEntry
}

var (
	client	*godis.Client
	blog	BlogStruct = BlogStruct{Name:"Stanley's Blog", Author:"stanley", Email:"stanley.w.cai@gmail.com"}
)

const (
	blogPrefix = "/blog/view/"
)

func (b *BlogStruct) add(entry BlogEntry) {
	log.Println("entry %v", entry)
	key := strconv.FormatInt(time.Now().Unix(), 10)
	value, err := json.Marshal(entry)
	log.Printf("entry %v value %v\n", entry, value)
	if err != nil {
		// log error msg and return
		log.Printf("failed to marshal(entry %v)\n", entry)
		return
	}

	client.Hset("blog", key, value)

	// TODO This implementation is dirty. Optimize this.
	// reload all blog entries
	b.loadAll()
}

func addNewEntries() {
	client.Flushdb()

	entry := BlogEntry{"hello, world", []byte("hello, go world"), time.Now().Unix()}
	blog.add(entry)
	entry = BlogEntry{"hello again", []byte("hello, GO world"), time.Now().Unix()}
	blog.add(entry)
	log.Println(blog)
}

func (b *BlogStruct) loadAll() {
	actual, err := client.Hkeys("blog")
	if err != nil {
		// log error msg and return
		log.Printf("failed to invoke Hkeys(blog) %v\n", actual)
		return
	}

	keys := make([]int64, len(actual))
	for i, key := range actual {
		keys[i], err = strconv.ParseInt(key, 10, 0)
		if err != nil {
			// log err message and continue
			log.Printf("failed to convert key(%v) into int64\n", key)
			keys[i] = 0
		}
	}

	// TODO: need sort the keys at first
	b.Entries = make([]BlogEntry, len(keys))
	for i, strKey := range actual {
		// JSON string to blogStruct
		value, err := client.Hget("blog", strKey)
		if err != nil {
			// log error message and continue
			log.Printf("failed to hget(blog, %v)\n", strKey)
			continue
		}
		err = json.Unmarshal(value.Bytes(), &b.Entries[i])
		if err != nil {
			// log error message and continue
			log.Printf("failed to unmarshal JSON %v\n", value)
		}
		log.Printf("value %v entry %v", value, b.Entries[i])
	}
	// Done
}

func blogIndex(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if _, ok := recover().(error); ok {
			// go error.html
			log.Println("failed")
		}
	}()
	if blog.Entries == nil {
		blog.loadAll()
	}
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, blog)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len(blogPrefix):]
	if blog.Entries == nil {
		blog.loadAll()
	}

	p, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		p = 0
	}
	var entry BlogEntry
	for _, e := range blog.Entries {
		if e.Date == p {
			entry = e
			break
		}
	}

	t, _ := template.ParseFiles("view.html")
	t.Execute(w, entry)
}

func main() {
	client = godis.New("", 0, "")
	addNewEntries()
	http.HandleFunc("/blog/", blogIndex)
	http.HandleFunc("/blog/view/", viewHandler)
	http.ListenAndServe(":8080", nil)
}
