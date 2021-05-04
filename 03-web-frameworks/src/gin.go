package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-jakarta/slides/03-web-frameworks/src/models"
	_ "github.com/mattn/go-sqlite3"
)

var db models.XODB

func init() {
	var err error

	models.XOLog = func(s string, p ...interface{}) {
		fmt.Printf("> SQL: %s -- %v\n", s, p)
	}

	// open database
	db, err = sql.Open("sqlite3", "db.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	r := gin.Default()
	r.GET("/hello/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.String(http.StatusOK, "hello %s", name)
	})

	r.GET("/authors", func(c *gin.Context) {
		// get authors
		authors, err := models.GetAuthors(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"authors": authors,
		})
	})

	// SPLIT OMIT

	r.POST("/add", func(c *gin.Context) {
		name := c.PostForm("name")
		author := &models.Author{
			Name: name,
		}

		// save to database
		err := author.Save(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id": author.AuthorID,
		})
	})

	r.Static("/static", "static")

	r.Run(":8000")
}
