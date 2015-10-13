package main

import (
	"encoding/xml"
	"github.com/slowteetoe/todo/Godeps/_workspace/src/github.com/gorilla/context"
	"github.com/slowteetoe/todo/Godeps/_workspace/src/gopkg.in/mgo.v2"
	"github.com/slowteetoe/todo/Godeps/_workspace/src/gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"os"
	"time"
)

type SmsResponse struct {
	XMLName xml.Name
	Message string `xml:"Sms"`
}

type TodoItem struct {
	Title       string
	CompletedAt *time.Time
}

type TodoList struct {
	Name      string
	TodoItems []TodoItem
}

func create(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db").(*mgo.Session)
	c := db.DB("").C("todolists")
	err := c.Insert(&TodoList{Name: "Basic Todo", TodoItems: []TodoItem{TodoItem{Title: "Eat Breakfast"}, TodoItem{Title: "Eat Lunch"}}})

	if err != nil {
		log.Fatal(err)
	}
}

func respond(w http.ResponseWriter, r *http.Request) {
	message := SmsResponse{XMLName: xml.Name{Local: "Response"}, Message: "Thank you, I got it."}
	x, err := xml.MarshalIndent(message, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write(x)
}

func todoList(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db").(*mgo.Session)
	c := db.DB("").C("todolists")
	result := TodoList{}
	err := c.Find(bson.M{"name": "Basic Todo"}).One(&result)
	if err != nil {
		log.Fatal(err)
	}
	x, err := xml.MarshalIndent(result, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write(x)
}

type Server struct {
	dbsession *mgo.Session
}

func NewMongoServer(connectionString string) (*Server, error) {
	dbsession, err := mgo.Dial(connectionString)
	if err != nil {
		return nil, err
	}
	return &Server{dbsession: dbsession}, nil
}

func (s *Server) Close() {
	s.dbsession.Close()
}

func (s *Server) WithData(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbcopy := s.dbsession.Copy()
		defer dbcopy.Close()
		context.Set(r, "db", dbcopy)
		fn(w, r)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("$PORT was unset, defaulting to %v", port)
	}
	mongoConnectionString := os.Getenv("MONGO_HOST")
	if mongoConnectionString == "" {
		mongoConnectionString = "localhost"
		log.Printf("$MONGO_HOST was unset, defaulting to %v", mongoConnectionString)
	}
	server, err := NewMongoServer(mongoConnectionString)
	if err != nil {
		panic(err)
	}
	defer server.Close()

	s := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	http.HandleFunc("/incoming", respond)
	http.HandleFunc("/list", server.WithData(todoList))
	http.HandleFunc("/create", server.WithData(create))

	log.Fatal(s.ListenAndServe())
}
