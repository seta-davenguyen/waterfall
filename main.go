package main

import (
	//"fmt"
	"log"
	//"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	//"github.com/araddon/dateparse"

	//"github.com/gorilla/context"
	//"github.com/julienschmidt/httprouter"
	//"github.com/justinas/alice"
	//"github.com/garyburd/redigo"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//https://github.com/gin-gonic/gin
//https://gist.github.com/congjf/8035830
//http://www.blog.labouardy.com/build-restful-api-in-go-and-mongodb/
func main() {

	router := gin.Default()

	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	data := &CommonData{session}

	router.GET("/auth/login/:name/:pass", data.loginHandle)
	router.GET("/auth/products/:idcat/:limit/:skip", data.getProducts)
	router.GET("/auth/productcs/:idcat/:limit/:skip", data.getProductCs)
	router.POST("/auth/addNewItem", data.addNewItem)
	router.POST("/auth/receiveItem", data.receiveItem)
	router.POST("/auth/deleteItem", data.deleteItem)
	router.POST("/auth/modifyItem", data.modifyItem)
	router.POST("/auth/completeItem", data.completeItem)

	router.Run(":8888")

}

type CommonData struct {
	session *mgo.Session
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

func (data *CommonData) loginHandle(c *gin.Context) {

	ss := data.session.Copy()
	defer ss.Close()

	users := ss.DB("waterfall").C("user")
	user := User{}

	log.Println("user info ", c.Param("name"), c.Param("pass"))

	err := users.Find(bson.M{"name": c.Param("name"), "pass": c.Param("pass")}).One(&user)

	if err != nil {
		log.Println("err:", err)
		c.JSON(400, gin.H{
			"error": true,
		})
	} else {
		log.Println("userall:", user)
		c.JSON(200, gin.H{
			"error": false,
			"id":    user.ID,
			"name":  user.Name,
			"role":  user.Role,
		})
	}
}
func (data *CommonData) getProducts(c *gin.Context) {
	ss := data.session.Copy()
	defer ss.Close()

	items := ss.DB("waterfall").C("items")

	var allItems []Item
	log.Println("idcat getProducts", c.Param("idcat"))
	limit, _ := strconv.Atoi(strings.Replace(c.Params.ByName("limit"), ":", "", -1))
	skip, _ := strconv.Atoi(strings.Replace(c.Params.ByName("skip"), ":", "", -1))
	idcat, _ := strconv.Atoi(strings.Replace(c.Params.ByName("idcat"), ":", "", -1))

	log.Println("idcat int ", idcat, limit, skip)
	//err := items.Find(bson.M{"idcat": 1}).All(&allItems)

	err := items.Find(bson.M{"idcat": idcat, "status": bson.M{"$lt": 4}}).Sort("status", "createddate").Limit(limit).Skip(skip).All(&allItems)
	//err := items.Find(nil).Limit(limit).Skip(skip).All(&allItems)

	if err != nil {
		log.Println("err getProducts:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": true,
		})
	} else {
		log.Println("size getProducts:", len(allItems))
		c.JSON(http.StatusOK, gin.H{
			"error": false,
			"items": allItems,
		})
	}
}
func (data *CommonData) getProductCs(c *gin.Context) {
	ss := data.session.Copy()
	defer ss.Close()

	items := ss.DB("waterfall").C("completed")

	var allItems []Completed
	log.Println("idcat getProducts", c.Param("idcat"))
	limit, _ := strconv.Atoi(strings.Replace(c.Params.ByName("limit"), ":", "", -1))
	skip, _ := strconv.Atoi(strings.Replace(c.Params.ByName("skip"), ":", "", -1))
	idcat, _ := strconv.Atoi(strings.Replace(c.Params.ByName("idcat"), ":", "", -1))

	log.Println("idcat int ", idcat, limit, skip)

	err := items.Find(nil).Sort("timec").Limit(limit).Skip(skip).All(&allItems)

	if err != nil {
		log.Println("err getProducts:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": true,
		})
	} else {
		log.Println("size getProducts:", len(allItems))
		c.JSON(http.StatusOK, gin.H{
			"error": false,
			"items": allItems,
		})
	}
}

func (data *CommonData) receiveItem(c *gin.Context) {

	item := Item{}

	err := c.BindJSON(&item)
	if err != nil {
		log.Println("err:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": true,
		})
		return
	}

	ss := data.session.Copy()
	defer ss.Close()

	items := ss.DB("waterfall").C("items")

	oldItem := Item{}

	err = items.FindId(bson.ObjectId(item.ID)).One(&oldItem)

	log.Println(item)

	log.Println("oldItem ", oldItem)

	//check this item was received or not
	if oldItem.Executor != "" && oldItem.Executor != item.Executor {
		log.Println("Don hang da duoc nhan ")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": true,
			"code":  1,
			"msg":   "Don hang da duoc nhan",
		})
		return
	}

	// Update
	matcher := bson.M{"_id": item.ID}
	change := bson.M{"$set": bson.M{"executor": item.Executor, "status": item.Status + 1}}
	err = items.Update(matcher, change)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": true,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error": false,
		})
	}
}

func (data *CommonData) completeItem(c *gin.Context) {

	comp := Completed{}

	err := c.BindJSON(&comp)
	if err != nil {
		log.Println("err:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": true,
		})
		return
	}

	ss := data.session.Copy()
	defer ss.Close()

	comps := ss.DB("waterfall").C("completed")

	oldItem := Item{}

	err = ss.DB("waterfall").C("items").FindId(bson.ObjectId(comp.ItemID)).One(&oldItem)

	log.Println(comp)
	log.Println("oldItem ", oldItem)

	//check this item was received or not
	//	if oldItem.Executor != "" && oldItem.Executor != item.Executor {
	//		log.Println("Don hang da duoc nhan ")
	//		c.JSON(http.StatusInternalServerError, gin.H{
	//			"error": true,
	//			"code":  1,
	//			"msg":   "Don hang da duoc nhan",
	//		})
	//		return
	//	}

	//add completed data to collection completed
	comp.Item = oldItem
	comp.TimeC = time.Now()
	comp.ID = bson.NewObjectId()
	err = comps.Insert(&comp)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": true,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error": false,
		})
	}

	//update status
	update := bson.M{"$set": bson.M{"status": oldItem.Status + 1}}
	ss.DB("waterfall").C("items").UpdateId(bson.ObjectId(oldItem.ID), update)

}

func (data *CommonData) deleteItem(c *gin.Context) {

	item := Item{}

	err := c.BindJSON(&item)
	if err != nil {
		log.Println("err:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": true,
		})
		return
	}

	ss := data.session.Copy()
	defer ss.Close()

	items := ss.DB("waterfall").C("items")

	oldItem := Item{}

	err = items.FindId(bson.ObjectId(item.ID)).One(&oldItem)
	if err != nil {
		log.Println("err:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": true,
		})
		return
	}

	log.Println(item)

	log.Println("oldItem ", oldItem)

	// Update
	matcher := bson.M{"_id": item.ID}
	change := bson.M{"$set": bson.M{"status": 5}}
	err = items.Update(matcher, change)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": true,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error": false,
		})
	}
}

func (data *CommonData) modifyItem(c *gin.Context) {

	item := Item{}

	err := c.BindJSON(&item)
	if err != nil {
		log.Println("err:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": true,
		})
		return
	}

	ss := data.session.Copy()
	defer ss.Close()

	items := ss.DB("waterfall").C("items")

	oldItem := Item{}

	err = items.FindId(bson.ObjectId(item.ID)).One(&oldItem)
	if err != nil {
		log.Println("err:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": true,
		})
		return
	}

	log.Println(item)

	log.Println("oldItem ", oldItem)

	// Update
	matcher := bson.M{"_id": item.ID}
	change := bson.M{"$set": bson.M{"cusName": item.CusName, "address": item.Address, "mobile": item.Mobile, "note": item.Note, "model": item.Model}}
	err = items.Update(matcher, change)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": true,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error": false,
		})
	}
}

func (data *CommonData) addNewItem(c *gin.Context) {

	item := Item{}

	//	var itemdata interface{}
	//	err := c.ShouldBind(&itemdata)
	//	if err != nil {
	//		log.Println("err should bind:", err)
	//		c.JSON(http.StatusBadRequest, gin.H{
	//			"error": true,
	//		})
	//		return
	//	}
	//	log.Println("itemdata:", itemdata)

	err := c.BindJSON(&item)
	if err != nil {
		log.Println("err:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": true,
		})
		return
	}

	ss := data.session.Copy()
	defer ss.Close()

	items := ss.DB("waterfall").C("items")
	item.ID = bson.NewObjectId()
	item.Status = 1
	err = items.Insert(&item)

	if err != nil {
		log.Println("err:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": true,
		})
		return
	}
	log.Println("item:", item)
	c.JSON(http.StatusOK, gin.H{
		"error": false,
		"id":    item.ID,
	})

}

//session, err := mgo.Dial("localhost")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer session.Close()
//	db := session.DB("dbname")

//	collection := db.C("property")
//	pipeLine := []m{
//		m{"$project": m{"address": 1, "st": 1, "city": 1, "notecount": m{"$size": "$notes"}}},  // output address, city, st, notecount
//		m{"$match": m{"notecount": m{"$gt": 2}}},                                               // keep docs with more than 2 notes
//		m{"$sort": bson.D{{"st", 1}, {"city", 1}}},                                             // sort results by state, city - see note above
//	}
//	iter := collection.Pipe(pipeLine).Iter()
//	defer iter.Close()

//	var result struct {
//		Id        string `bson:"_id"`
//		State     string `bson:"st"`
//		City      string `bson:"city"`
//		Address   string `bson:"address"`
//		NoteCount int    `bson:"notecount"`
//	}
//	for iter.Next(&result) {
//		log.Printf("%+v", result)
//	}
//	if iter.Err() != nil {
//		log.Println(iter.Err())
//	}
// Find book which prices is greater than(gt) 40

//	iter := c.Find(bson.M{"price": bson.M{"$gt": 40}}).Iter()
//	var index = 1

//	for iter.Next(&result) {
//		fmt.Printf("current result is [%d] result =%+v\n", index, result)
//		index++
//	}
//	//when search the DB it must all lower-case to avoid any error.
//	if err2 := iter.Close(); err2 != nil {
//		fmt.Printf("No data\n")
//	} else {
//		fmt.Printf("result =%+v\n", result)
//	}
