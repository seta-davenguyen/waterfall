package main

import (
	"log"
	"net/http"
	"time"

	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func CreateItemIndex(session *mgo.Session) {

	ss := session.Copy()
	defer ss.Close()

	itemIndex := mgo.Index{
		Key:        []string{"idcat", "status", "createddate"},
		Unique:     false,
		Background: true,
		Sparse:     false,
	}

	items := ss.DB("waterfall").C("item")

	err := items.EnsureIndex(itemIndex)
	if err != nil {
		log.Fatalf("[CreateItemIndex] - create index failed %s\n", err)
	}

	completedIndex := mgo.Index{
		Key:        []string{"timec"},
		Unique:     false,
		Background: true,
		Sparse:     false,
	}

	compItem := ss.DB("waterfall").C("completed")

	err = compItem.EnsureIndex(completedIndex)
	if err != nil {
		log.Fatalf("[CreateItemIndex] - create index failed %s\n", err)
	}
}

func GetProducts(session *mgo.Session) gin.HandlerFunc {
	log.Println("GetProducts")
	return func(c *gin.Context) {
		ss := session.Copy()
		defer ss.Close()

		items := ss.DB("waterfall").C("items")

		claimInfo, exist := c.Get("claimInfo")
		if exist {
			log.Println("claiminfo", claimInfo.(*ClaimInfo))
		}

		var allItems []Item
		log.Println("limit getProducts", c.Params.ByName("limit"))

		limit, err := strconv.Atoi(c.Params.ByName("limit"))
		if err != nil || limit > LIMIT_MAX {
			log.Println("size getProducts:", allItems)
			c.JSON(http.StatusOK, gin.H{
				"error": INVALID_PARAM,
			})
		}
		skip, err := strconv.Atoi(c.Params.ByName("skip"))
		if err != nil {
			log.Println("size getProducts:", allItems)
			c.JSON(http.StatusOK, gin.H{
				"error": INVALID_PARAM,
			})
		}
		idcat, err := strconv.Atoi(c.Params.ByName("idcat"))
		if err != nil {
			log.Println("size getProducts:", allItems)
			c.JSON(http.StatusOK, gin.H{
				"error": INVALID_PARAM,
			})
		}

		err = items.Find(bson.M{"idcat": idcat, "status": bson.M{"$lt": 4}}).Sort("status", "createddate").Limit(limit).Skip(skip).All(&allItems)
		//err := items.Find(nil).Limit(limit).Skip(skip).All(&allItems)

		if err != nil {
			log.Println("err getProducts:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": NO_DOC,
			})
		} else {
			log.Println("size getProducts:", allItems)
			c.JSON(http.StatusOK, gin.H{
				"error": NO_ERROR,
				"items": allItems,
			})
		}
	}
}

func GetProductCs(session *mgo.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		ss := session.Copy()
		defer ss.Close()

		items := ss.DB("waterfall").C("completed")

		var allItems []Completed
		limit, err := strconv.Atoi(c.Params.ByName("limit"))
		if err != nil || limit > LIMIT_MAX {
			log.Println("size getProducts:", allItems)
			c.JSON(http.StatusOK, gin.H{
				"error": INVALID_PARAM,
			})
		}
		skip, err := strconv.Atoi(c.Params.ByName("skip"))
		if err != nil {
			log.Println("size getProducts:", allItems)
			c.JSON(http.StatusOK, gin.H{
				"error": INVALID_PARAM,
			})
		}
		idcat, err := strconv.Atoi(c.Params.ByName("idcat"))
		if err != nil {
			log.Println("size getProducts:", allItems)
			c.JSON(http.StatusOK, gin.H{
				"error": INVALID_PARAM,
			})
		}

		err = items.Find(nil).Sort("timec").Limit(limit).Skip(skip).All(&allItems)

		if err != nil {
			log.Println("err getProducts:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": NO_DOC,
			})
		} else {
			log.Println("size getProducts:", len(allItems))
			c.JSON(http.StatusOK, gin.H{
				"error": NO_ERROR,
				"items": allItems,
			})
		}
	}
}

func AddNewItem(session *mgo.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		item := Item{}

		err := c.BindJSON(&item)
		if err != nil {
			log.Println("err:", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": ER_PARSOR,
			})
			return
		}

		ss := session.Copy()
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
			"error": NO_ERROR,
			"id":    item.ID,
		})
	}
}

func ReceiveItem(session *mgo.Session) gin.HandlerFunc {
	return func(c *gin.Context) {

		item := Item{}

		err := c.BindJSON(&item)
		if err != nil {
			log.Println("err:", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": ER_PARSOR,
			})
			return
		}

		ss := session.Copy()
		defer ss.Close()

		log.Println("item ", item)

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
				"error": NO_DOC,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"error": false,
			})
		}
	}
}
func CompleteItem(session *mgo.Session) gin.HandlerFunc {
	return func(c *gin.Context) {

		comp := Completed{}

		err := c.BindJSON(&comp)
		if err != nil {
			log.Println("err:", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": ER_PARSOR,
			})
			return
		}

		ss := session.Copy()
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
}

func DeleteItem(session *mgo.Session) gin.HandlerFunc {
	return func(c *gin.Context) {

		item := Item{}

		err := c.BindJSON(&item)
		if err != nil {
			log.Println("err:", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": ER_PARSOR,
			})
			return
		}

		ss := session.Copy()
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
}

func MoveItem(session *mgo.Session) gin.HandlerFunc {
	return func(c *gin.Context) {

		item := Item{}

		err := c.BindJSON(&item)
		if err != nil {
			log.Println("err:", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": ER_PARSOR,
			})
			return
		}

		ss := session.Copy()
		defer ss.Close()

		log.Println(item)

		items := ss.DB("waterfall").C("items")

		//	oldItem := Item{}

		//	err = items.FindId(bson.ObjectId(item.ID)).One(&oldItem)
		//	if err != nil {
		//		log.Println("err:", err)
		//		c.JSON(http.StatusBadRequest, gin.H{
		//			"error": true,
		//		})
		//		return
		//	}

		//log.Println("oldItem ", oldItem)

		// Update
		matcher := bson.M{"_id": item.ID}
		change := bson.M{"$set": bson.M{"idcat": 2}}
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
}
func ModifyItem(session *mgo.Session) gin.HandlerFunc {
	return func(c *gin.Context) {

		item := Item{}

		err := c.BindJSON(&item)
		if err != nil {
			log.Println("err:", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": ER_PARSOR,
			})
			return
		}

		ss := session.Copy()
		defer ss.Close()

		items := ss.DB("waterfall").C("items")

		oldItem := Item{}

		err = items.FindId(bson.ObjectId(item.ID)).One(&oldItem)
		if err != nil {
			log.Println("err:", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": NO_DOC,
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
				"error": NO_DOC,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"error": false,
			})
		}
	}
}
