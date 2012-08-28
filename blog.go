package main

import (
	"encoding/json"
	"github.com/simonz05/godis"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

type BlogEntry struct {
	Title string
	Body  []byte
	Date  int64
}

type BlogStruct struct {
	Name    string
	Author  string
	Email   string
	Entries []BlogEntry
}

var (
	client *godis.Client
	blog   BlogStruct = BlogStruct{Name: "Stanley's Blog", Author: "stanley", Email: "stanley.w.cai@gmail.com"}
)

const (
	blogPrefix = "/blog/view/"
)

func (b *BlogStruct) add(entry BlogEntry) error {
	key := strconv.FormatInt(time.Now().Unix(), 10)
	value, err := json.Marshal(entry)
	// log.Printf("entry %v value %v\n", entry, value)
	if err != nil {
		return err
	}

	if ok, err := client.Hset("blog", key, value); !ok && err != nil {
		return err
	}

	// TODO This implementation is dirty. Optimize this.
	// reload all blog entries
	return b.loadAll()
}

func (b *BlogStruct) addNewEntries() error {
	// clean up the database at first
	client.Flushdb()

	entry := BlogEntry{"hello, world", []byte("hello, go world"), time.Now().Unix()}
	if err := b.add(entry); err != nil {
		return err
	}

	// sleep 2 seconds to make sure two different items are added
	time.Sleep(1 * 1e9)
	entry = BlogEntry{"hello again", []byte("hello, GO world"), time.Now().Unix()}
	if err := b.add(entry); err != nil {
		return err
	}
	return nil
}

func (b *BlogStruct) loadAll() error {
	actual, err := client.Hkeys("blog")
	if err != nil {
		// log error msg and return
		// log.Printf("failed to invoke Hkeys(blog) %v\n", actual)
		return err
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
	}
	// Done
	return nil
}

func blogIndex(w http.ResponseWriter, r *http.Request) error {
	defer func() {
		if _, ok := recover().(error); ok {
			// go error.html
			log.Println("failed")
		}
	}()
	if blog.Entries == nil {
		if err := blog.loadAll(); err != nil {
			return err
		}
	}
	t, err := template.ParseFiles("index.html")
	if err != nil {
		return err
	}
	return t.Execute(w, blog)
}

func viewBlogEntry(w http.ResponseWriter, r *http.Request) error {
	id := r.URL.Path[len(blogPrefix):]
	if blog.Entries == nil {
		if err := blog.loadAll(); err != nil {
			return err
		}
	}

	p, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		p = 0
		return err
	}

	var entry BlogEntry
	for _, e := range blog.Entries {
		if e.Date == p {
			entry = e
			break
		}
	}

	t, err := template.ParseFiles("view.html")
	if err != nil {
		return err
	}
	return t.Execute(w, entry)
}

type appHandler func(http.ResponseWriter,*http.Request) error

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func main() {
	client = godis.New("", 0, "")
	blog.addNewEntries()
	http.Handle("/blog/", appHandler(blogIndex))
	http.Handle("/blog/view/", appHandler(viewBlogEntry))
	http.ListenAndServe(":8080", nil)
}
