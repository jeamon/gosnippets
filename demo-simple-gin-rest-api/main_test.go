package main

// Basic test file for <demo-simple-gin-rest-api.go> snippet.

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBooksHandlers(t *testing.T) {
	mockStore := &BooksStore{
		Books: []Book{
			{ID: "1", Name: "Computer Science", Author: "Jerome"},
		},
	}
	gin.SetMode(gin.TestMode)
	testServer := httptest.NewServer(mockStore.SetRoutes(gin.Default()))
	defer testServer.Close()

	t.Run("Testing Books Handler", func(t *testing.T) {
		t.Run("GetAllBooks: should pass", func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/books/", nil)
			assert.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))
			assert.NoError(t, err)
			expected, err := json.Marshal(gin.H{
				"books": mockStore.Books,
			})
			assert.NoError(t, err)
			assert.Equal(t, expected, body)
		})

		t.Run("GetBook: found case", func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/books/1", nil)
			assert.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))
			assert.NoError(t, err)
			expected, err := json.Marshal(gin.H{
				"message": "found",
				"book":    mockStore.Books[0],
			})
			assert.NoError(t, err)
			assert.Equal(t, expected, body)
		})

		t.Run("GetBook: not found case", func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/books/0", nil)
			assert.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
			assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))
			assert.NoError(t, err)
			expected, err := json.Marshal(gin.H{
				"message": "not found",
			})
			assert.NoError(t, err)
			assert.Equal(t, expected, body)
		})
	})
}

/*

~$ go test -v
=== RUN   TestBooksHandlers
=== RUN   TestBooksHandlers/Testing_Books_Handler
=== RUN   TestBooksHandlers/Testing_Books_Handler/GetALLBooks:_should_pass
[GIN] 2022/07/07 - 02:36:54 | 200 |            0s |       127.0.0.1 | GET      "/api/v1/books/"
=== RUN   TestBooksHandlers/Testing_Books_Handler/GetBook:_found_case
[GIN] 2022/07/07 - 02:36:54 | 200 |            0s |       127.0.0.1 | GET      "/api/v1/books/1"
=== RUN   TestBooksHandlers/Testing_Books_Handler/GetBook:_not_found_case
[GIN] 2022/07/07 - 02:36:54 | 404 |            0s |       127.0.0.1 | GET      "/api/v1/books/0"
--- PASS: TestBooksHandlers (0.00s)
    --- PASS: TestBooksHandlers/Testing_Books_Handler (0.00s)
        --- PASS: TestBooksHandlers/Testing_Books_Handler/GetALLBooks:_should_pass (0.00s)
        --- PASS: TestBooksHandlers/Testing_Books_Handler/GetBook:_found_case (0.00s)
        --- PASS: TestBooksHandlers/Testing_Books_Handler/GetBook:_not_found_case (0.00s)
PASS
ok      github.com/jeamon/live-demo-books-store 0.392s

*/
