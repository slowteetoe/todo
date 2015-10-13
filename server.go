package main

import (
	"encoding/xml"
	"github.com/slowteetoe/todo/Godeps/_workspace/src/github.com/gorilla/context"
	"github.com/slowteetoe/todo/Godeps/_workspace/src/gopkg.in/mgo.v2"
	"github.com/slowteetoe/todo/Godeps/_workspace/src/gopkg.in/mgo.v2/bson"
	"io/ioutil"
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
	Id                     bson.ObjectId `bson:"_id,omitempty"`
	Name                   string
	TodoItems              []TodoItem
	AssociatedPhoneNumbers []string
}

func create(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db").(*mgo.Session)
	c := db.DB("").C("todolists")

	todoList, err := createBlankTodoList("+17022181502", "Basic Todo", c)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Created %v", todoList)
}

func createBlankTodoList(phoneNumber, todoListName string, c *mgo.Collection) (*TodoList, error) {
	var todoList TodoList
	todoList = TodoList{
		Id:                     bson.NewObjectId(),
		Name:                   todoListName,
		TodoItems:              []TodoItem{},
		AssociatedPhoneNumbers: []string{phoneNumber},
	}
	err := c.Insert(&todoList)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &todoList, nil
}

// Twilio sends a POST to the specified endpoint
// e.g.
// ToCountry=US&ToState=NV&SmsMessageSid=SM1b9d6ec899fc86c6c08b73e1bfb7861c&NumMedia=0&ToCity=&FromZip=89150&SmsSid=SM1b9d6ec899fc86c6c08b73e1bfb7861c&FromState=NV&SmsStatus=received&FromCity=LAS+VEGAS&Body=Test&FromCountry=US&To=%2B17025000247&ToZip=&NumSegments=1&MessageSid=SM1b9d6ec899fc86c6c08b73e1bfb7861c&AccountSid=AC3bcd52a18af4c60d5d63d3408973f830&From=%2B17022181502&ApiVersion=2010-04-01
func incoming(w http.ResponseWriter, r *http.Request) {
	body := r.PostFormValue("Body")
	from := r.PostFormValue("From")

	if body == "" || from == "" {
		http.Error(w, "Missing required key", http.StatusInternalServerError)
		data, _ := ioutil.ReadAll(r.Body)
		log.Printf("Unable to read expected fields from %v", string(data))
		return
	}

	db := context.Get(r, "db").(*mgo.Session)
	c := db.DB("").C("todolists")
	todo, err := todoListFor(from, c)

	if err != nil {
		log.Printf("Error getting todolist: %v", err)
	}

	if todo == nil {
		log.Printf("Didn't pull up a todo list, creating a new one...")
		todo, err = createBlankTodoList(from, "Basic Todo", c)
		if err != nil {
			log.Printf("Creating a new blank todo didn't work right: %v", err)
		}
	}

	log.Printf("Working with %v", todo)
	todo.TodoItems = append(todo.TodoItems, TodoItem{Title: body})
	log.Printf("TODO: need to save %v", todo)
	// c.UpdateId(todo.Id, bson.M{"$set": bson.M{"name": "updated name"}})

	message := SmsResponse{XMLName: xml.Name{Local: "Response"}, Message: "Thank you, I got it."}
	x, err := xml.MarshalIndent(message, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write(x)
}

func todoListFor(phoneNumber string, c *mgo.Collection) (*TodoList, error) {
	result := TodoList{}
	err := c.Find(bson.M{"name": "Basic Todo"}).One(&result)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &result, nil
}

func todoList(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db").(*mgo.Session)
	c := db.DB("").C("todolists")
	todoList, err := todoListFor("", c)
	x, err := xml.MarshalIndent(todoList, "", "  ")
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
	http.HandleFunc("/incoming", server.WithData(incoming))
	http.HandleFunc("/list", server.WithData(todoList))
	http.HandleFunc("/create", server.WithData(create))

	log.Fatal(s.ListenAndServe())
}
