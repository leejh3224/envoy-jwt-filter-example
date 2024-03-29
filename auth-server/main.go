package main

import (
	"crypto/x509"
	"encoding/pem"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"io/ioutil"
	"net/http"
	"time"
)

type getTokenRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type user struct {
	Username string
	Password string
	Role     string
}

func main() {
	r := gin.Default()
	r.Use(cors.Default())

	// Endpoint for retrieving JWT
	// Use admin for username and 1234 for password to get token
	r.POST("/token", func(c *gin.Context) {
		var body getTokenRequest
		_ = c.BindJSON(&body)

		users := []user{
			{Username: "admin", Password: "1234", Role: "admin"},
			{Username: "user-1234", Password: "1234", Role: "user"},
		}

		var userFound *user
		for _, user := range users {
			if body.Username == user.Username && body.Password == user.Password {
				userFound = &user
				break
			}
		}

		if userFound == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username or password"})
			return
		}

		rawKey, err := ioutil.ReadFile("./private_key.rsa")
		if err != nil {
			panic(err)
		}

		block, _ := pem.Decode(rawKey)
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			panic(err)
		}

		unsigned, err := jwt.NewBuilder().
			Issuer("auth-server").
			IssuedAt(time.Now()).
			Subject(userFound.Role).
			Expiration(time.Now().AddDate(0, 0, 7)).
			Build()
		if err != nil {
			panic(err)
		}

		signed, err := jwt.Sign(unsigned, jwt.WithKey(jwa.RS256, key))
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, gin.H{"token": string(signed)})
	})

	// Endpoint for retrieving JSON Web Keys (jwks)
	// For more information, please see https://datatracker.ietf.org/doc/html/rfc7517
	r.GET("/.well-known/jwks.json", func(c *gin.Context) {
		data, err := ioutil.ReadFile("./private_key.rsa")
		if err != nil {
			panic(err)
		}

		key, err := jwk.ParseKey(data, jwk.WithPEM(true))
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, gin.H{"keys": []jwk.Key{key}})
	})

	r.Run(":8081")
}
