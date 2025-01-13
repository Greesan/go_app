// main_test.go

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"time"
)

func TestCreateBook(t *testing.T) {
	if err := InitDB(); err != nil {
        t.Fatalf("Failed to initialize database: %v", err)
    }
    defer DB.Close()
	router := gin.Default()
	router.POST("/books", CreateBook)
	// Create a test book
	book := Book{
		Title:           "Test Book",
		Summary:         "A test summary",
		Author:          "Test Author",
		First_published: time.Date(2025, 1, 9, 16, 0, 0, 0, time.FixedZone("PST", -8*60*60)),
	}

	// Convert book to JSON
	jsonBook, _ := json.Marshal(book)

	// Create a request
	req, _ := http.NewRequest("POST", "/books", bytes.NewBuffer(jsonBook))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, 200, w.Code)

	// You might want to check the response body as well
	var responseBook Book
	json.Unmarshal(w.Body.Bytes(), &responseBook)
	assert.Equal(t, book.Title, responseBook.Title)
	// Add more assertions as needed
}