package main

import (
	"log"
	"testing"
)

var mongoServer *MongoServer

func init() {
	db, err := NewMongoServer("mongodb://localhost/todolist")
	mongoServer = db
	if err != nil {
		log.Printf("Unable to connect to db %v", err)
	}
}

func TestFindingTodoListByPhoneNumber(t *testing.T) {
	db := mongoServer.dbsession.Copy()
	defer db.Close()

	todo, err := todoListFor("+17022181502", db)
	if err != nil {
		t.Errorf("Unable to find todolist %v", err)
	}
	log.Printf("%v", todo)
}
