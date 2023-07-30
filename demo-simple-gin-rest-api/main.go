package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// This is a vanilla fun live coding challenge with below requirements :
// Design and Implement Rest API Endpoint to display all books.
// Design and Implement Rest API Endpoint to display specific book based on ID.
// Below is a simple quick-dirty solution I crafted for the purpose. Enjoy.

type Book struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Author string `json:"author"`
}

type BooksStore struct {
	Books []Book
}

func (store *BooksStore) SetRoutes(router *gin.Engine) *gin.Engine {
	api := router.Group("/api/v1/books")
	api.GET("/", store.GetAllBooks())
	api.GET("/:id", store.GetBook())
	return router
}

func main() {

	store := &BooksStore{
		Books: []Book{
			{ID: "1", Name: "Computer Science", Author: "Jerome"},
			{ID: "2", Name: "Algorithms", Author: "Jerome"},
			{ID: "3", Name: "Maths", Author: "Jerome"},
		},
	}

	router := gin.Default()
	store.SetRoutes(router)

	api := http.Server{
		Handler: router,
		Addr:    fmt.Sprintf("%s:%s", "localhost", "8080"),
	}

	if err := api.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}

func (s *BooksStore) GetAllBooks() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"books": s.Books,
		})
	}
}

func (s *BooksStore) GetBook() func(c *gin.Context) {
	return func(c *gin.Context) {

		id := c.Param("id")
		var book Book
		for _, b := range s.Books {
			if b.ID == id {
				book = b
			}
		}

		if book.ID == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "found",
			"book":    book,
		})
	}
}

// curl -X GET http://localhost:8080/api/v1/books
// curl -X GET http://localhost:8080/api/v1/books/1
// curl -X GET http://localhost:8080/api/v1/books/0
