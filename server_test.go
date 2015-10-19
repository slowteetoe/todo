package main

import (
	"log"
	"testing"
)

var db *MongoServer

func init() {
	server, err := NewMongoServer("mongodb://localhost/todolist")
	db = server
	if err != nil {
		log.Printf("Unable to connect to db %v", err)
	}
}

func TestFindingTodoListByPhoneNumber(t *testing.T) {
	db.dbsession.Copy()
	defer db.Close()

	c := db.dbsession.DB("").C("todolists")
	todo, err := todoListFor("+17022181502", c)
	if err != nil {
		t.Errorf("Unable to find todolist %v", err)
	}
	log.Printf("%v", todo)
}
