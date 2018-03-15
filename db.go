package main

import (
	"log"
	"time"

	"gopkg.in/mgo.v2"
)

func CreateSession() *mgo.Session {
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{"127.0.0.1:27017"},
		Username: "",
		Password: "",
		Timeout:  60 * time.Second,
	})

	if err != nil {
		log.Fatalf("[CreateSession] - error create session %s\n", err)
	}
	return session
}

func CreateIndex(session *mgo.Session) {
	CreateItemIndex(session)
	CreateUserIndex(session)
}
