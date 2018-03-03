package main

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"gopkg.in/mgo.v2/bson"
)

type ClaimInfo struct {
	ID   string
	Role int
	jwt.StandardClaims
}
type User struct {
	ID          bson.ObjectId `json:"id" bson:"_id"`
	Name        string        `json:"name"`
	Pass        string        `json:"pass"`
	Role        int           `json:"role"`
	CreatedDate string        `json:"createdDate"`
}
type Item struct {
	ID          bson.ObjectId `json:"id" bson:"_id"`
	IdCat       int32         `json:"idCat"`
	CusName     string        `json:"cusName"`
	Model       string        `json:"model"`
	Note        string        `json:"note"`
	Address     string        `json:"address"`
	Mobile      string        `json:"mobile"`
	Status      int32         `json:"status"`
	Adder       string        `json:"adder"`
	Executor    string        `json:"executor"`
	CreatedDate time.Time     `json:"createdDate" time_format:"2006-01-02" time_utc:"7"`
	ActionDate  time.Time     `json:"actionDate" time_format:"2006-01-02" time_utc:"7"`
}

type Completed struct {
	ID       bson.ObjectId `json:"id" bson:"_id"`
	ItemID   bson.ObjectId `json:"itemId"`
	Item     Item          `json:"item"`
	Price    float32       `json:"price"`
	Quantity int           `json:"quantity"`
	Revenue  float32       `json:"revenue"`
	ModelC   string        `json:"modelc"`
	NoteC    string        `json:"notec"`
	TimeC    time.Time     `json:"timec"`
}
