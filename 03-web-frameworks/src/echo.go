package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kenshaw/go-jakarta/03-web-frameworks/src/models"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
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
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/hello/:name", func(c echo.Context) error {
		name := c.Param("name")
		return c.String(http.StatusOK, fmt.Sprintf("hello %s", name))
	})

	e.GET("/authors", func(c echo.Context) error {
		// get authors
		authors, err := models.GetAuthors(db)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, gin.H{
			"authors": authors,
		})
	})

	// SPLIT OMIT

	e.POST("/add", func(c echo.Context) error {
		name := c.FormValue("name")
		author := &models.Author{
			Name: name,
		}

		// save to database
		err := author.Save(db)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, gin.H{
			"id": author.AuthorID,
		})
	})

	e.Static("/static", "static")

	e.Run(standard.New(":8000"))
}
