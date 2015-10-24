package main

import (
	"github.com/slowteetoe/todo/Godeps/_workspace/src/gopkg.in/mgo.v2"
	"github.com/slowteetoe/todo/Godeps/_workspace/src/gopkg.in/mgo.v2/bson"
	"log"
	"time"
)

type MongoServer struct {
	dbsession *mgo.Session
}

func NewMongoServer(connectionString string) (*MongoServer, error) {
	dbsession, err := mgo.Dial(connectionString)
	if err != nil {
		return nil, err
	}
	return &MongoServer{dbsession: dbsession}, nil
}

func (s *MongoServer) Close() {
	s.dbsession.Close()
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

func createBlankTodoList(phoneNumber, todoListName string, db *mgo.Session) (*TodoList, error) {
	var todoList TodoList
	todoList = TodoList{
		Id:                     bson.NewObjectId(),
		Name:                   todoListName,
		TodoItems:              []TodoItem{},
		AssociatedPhoneNumbers: []string{phoneNumber},
	}
	c := db.DB("").C("todolists")

	err := c.Insert(&todoList)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &todoList, nil
}

func todoListFor(phoneNumber string, db *mgo.Session) (*TodoList, error) {
	result := TodoList{}
	c := db.DB("").C("todolists")
	err := c.Find(bson.M{"name": "Basic Todo"}).One(&result)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &result, nil
}
