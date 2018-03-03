package main

import (
	"crypto/rsa"
	"io/ioutil"
	"log"
	"net/http"

	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// verify key and sign key
var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

// read the key files before starting http router
func initKeys() {
	signBytes, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		log.Fatalf("[init Keys]: %s\n", err)
	}

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatalf("[init Keys]: %s\n", err)
	}

	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		log.Fatalf("[init Keys]: %s\n", err)
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatalf("[init Keys]: %s\n", err)
	}
}

func CreateUserIndex(session *mgo.Session) {
	userIndex := mgo.Index{
		Key:        []string{"name"},
		Unique:     true,
		Background: true,
		Sparse:     false,
	}

	users := session.DB("waterfall").C("user")
	err := users.EnsureIndex(userIndex)
	if err != nil {
		log.Fatalln("[CreateUserIndex] - error create index ", err)
	}
}
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from request
		token, err := request.ParseFromRequestWithClaims(c.Request, request.OAuth2Extractor, &ClaimInfo{}, func(token *jwt.Token) (interface{}, error) {
			return verifyKey, nil
		})

		if err != nil {
			switch err.(type) {
			case *jwt.ValidationError: // JWT validation error
				vErr := err.(*jwt.ValidationError)
				switch vErr.Errors {
				case jwt.ValidationErrorExpired: //JWT expired
					c.JSON(http.StatusUnauthorized, gin.H{
						"error": TOKEN_EXPIRED,
					})
					log.Printf("[AuthRequired]- Token Expired, get a new one")
					c.Abort()
					return
				default:
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": ER_INTERNAL,
					})
					log.Printf("[AuthRequired]- ValidationError error: %+v\n", vErr.Errors)
					c.Abort()
					return
				}
			default:
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": NO_TOKEN,
				})
				log.Printf("[AuthRequired]- Token parse error: %v\n", err)
				c.Abort()
				return
			}
		}
		if token.Valid {
			// Set user name to HTTP context
			//set claimInfo for using in later router
			c.Set("claimInfo", token.Claims.(*ClaimInfo))

			//set userID for display user info in log of gin. Useful in debugging
			// Need to change the gin.Logger as in https://github.com/huyntsgs/gin/blob/huyntsgs-logger-info/logger.go
			c.Set("userID", token.Claims.(*ClaimInfo).ID)
			c.Next()
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": TOKEN_INVALID,
			})
			log.Printf("[AuthRequired]- Token is invalid")
			c.Abort()
			return
		}
	}
}

func LoginHandle(session *mgo.Session) gin.HandlerFunc {
	return func(c *gin.Context) {

		ss := session.Copy()
		defer ss.Close()

		users := ss.DB("waterfall").C("user")
		user := User{}
		userData := User{}

		err := c.BindJSON(&userData)

		if err != nil {
			log.Println("err:", err)
			c.JSON(400, gin.H{
				"error": ER_PARSOR,
			})
		}
		log.Println("user info ", userData)

		err = users.Find(bson.M{"name": userData.Name, "pass": userData.Pass}).One(&user)

		if err != nil {
			log.Println("err:", err)
			c.JSON(400, gin.H{
				"error": NO_DOC,
			})
		} else {

			claimInfo := ClaimInfo{user.ID.Hex(), user.Role,
				jwt.StandardClaims{
					ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
					Issuer:    "admin",
				}}
			token := jwt.NewWithClaims(jwt.SigningMethodRS256, claimInfo)
			tokenString, err := token.SignedString(signKey)
			if err != nil {
				log.Printf("Token Signing error: %v\n", err)
			}
			c.JSON(200, gin.H{
				"error": NO_ERROR,
				"id":    user.ID,
				"name":  user.Name,
				"role":  user.Role,
				"token": tokenString,
			})
		}
	}
}
