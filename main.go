package main

import (

	//"github.com/garyburd/redigo"

	"github.com/gin-gonic/gin"
	"google.golang.org/appengine"
	"gopkg.in/mgo.v2"
)

func init() {
	initKeys()
}

func main() {

	session := CreateSession()
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	CreateIndex(session)

	router := gin.New()
	//use LoggerInfo instead Logger to include userID in request log
	router.Use(gin.LoggerInfo("userID"))
	router.Use(gin.Recovery())

	router.POST("/login", LoginHandle(session))

	authorized := router.Group("/auth")
	authorized.Use(AuthRequired())
	{
		authorized.GET("/products/:idcat/:limit/:skip", GetProducts(session))
		authorized.GET("/productcs/:idcat/:limit/:skip", GetProductCs(session))
		authorized.POST("/addNewItem", AddNewItem(session))
		authorized.POST("/receiveItem", ReceiveItem(session))
		authorized.POST("/deleteItem", DeleteItem(session))
		authorized.POST("/modifyItem", ModifyItem(session))
		authorized.POST("/completeItem", CompleteItem(session))
		authorized.POST("/moveItem", MoveItem(session))
	}

	router.Run(":8888")
	appengine.Main() // required!

}
