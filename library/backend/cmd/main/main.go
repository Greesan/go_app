package main
import (
    "fmt"
    _ "github.com/go-sql-driver/mysql"
    "net/http"
	"github.com/gin-gonic/gin"
    "time"
    "log"
    "github.com/jmoiron/sqlx"
    "html/template"
    "net/url"
    "strconv"
    "github.com/rs/cors"

)

type Book struct {
    Book_id   int64 `json:"Book_id"`
    Title string `json:"Title"`
    Summary string `json:"Summary"`
    Author string`json:"Author"`
    First_published time.Time `json:"First_published,omitempty"`
    Last_updated time.Time `json:"Last_updated,omitempty"`
}

func (b Book) String() string {
    return fmt.Sprintf("Title: %s, Summary: %s, Author: %s, First_published: %s, Last_updated: %s", b.Title, b.Summary, b.Author, b.First_published, b.Last_updated)
}

var DB *sqlx.DB
var err error

func getBooks(c *gin.Context) {
	books, err := loadBooksFromDB(c)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK,books)
}


func CreateBook(c *gin.Context) {
    var book Book
    if err := c.ShouldBindJSON(&book); err != nil {
        log.Printf("Error binding JSON: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Parse the date string into a time.Time object
    if err != nil {
        log.Printf("Error parsing First_published date: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format for First_published"})
        return
    }
    
    // Use the current time for Last_updated
    lastUpdated := time.Now()
    result, err := DB.Exec("INSERT INTO books (Title, Summary, Author, First_published, Last_updated) VALUES (?, ?, ?, ?, ?)", book.Title, book.Summary, book.Author, book.First_published,lastUpdated)
    if err != nil {
        log.Printf("Error creating book: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    lastID, err := result.LastInsertId()
    if err != nil {
        log.Printf("Error getting last insert ID: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving book ID"})
        return
    }
    
    // Update the book object with the new ID and parsed dates
    book.Book_id = int64(lastID)
    book.Last_updated = lastUpdated
    
    c.JSON(http.StatusOK, book)
}

func CreateBookFromForm(c *gin.Context) {
    rawData, _ := c.GetRawData()
    fmt.Println("RawData:" + string(rawData))
    parsedData, err:= url.ParseQuery(string(rawData))
    if err != nil {
        fmt.Println("Error parsing query:", err)
        return
    }

    // Output the parsed parameters
    fmt.Printf("parsedData:")
    fmt.Println(parsedData)
    t, _ :=time.Parse("2006-01-02", parsedData.Get("First_published"))
    book := Book{
        Title: parsedData.Get("Title"),
        Author: parsedData.Get("Author"),
        Summary: parsedData.Get("Summary"),
        First_published: t,
    }
    fmt.Printf("Book after binding: %+v\n",book)

    // Use the current time for Last_updated
    lastUpdated := time.Now()
    fmt.Println("Running insert from Form:")
    result, err := DB.Exec("INSERT INTO books (Title, Summary, Author, First_published, Last_updated) VALUES (?, ?, ?, ?, ?)", 
        book.Title, book.Summary, book.Author, book.First_published, lastUpdated)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    lastID, _ := result.LastInsertId()
    book.Book_id = lastID
    book.Last_updated = lastUpdated

    c.Redirect(http.StatusSeeOther, "/books")
}

func InitDB() error {
    var err error
    DB, err = sqlx.Connect("mysql", "root:bricha101@tcp(127.0.0.1:3306)/library_system?parseTime=true")
    if err != nil {
        return fmt.Errorf("failed to connect to database: %v", err)
    }
    return nil
}

func showForm(c *gin.Context) {
    tmpl, err := template.ParseFiles("templates/form.html")
    if err != nil {
        c.String(http.StatusInternalServerError, err.Error())
        return
    }
    tmpl.Execute(c.Writer, nil)
}

func loadBooksFromDB(c *gin.Context) ([]Book, error) {
    // Query the database for books
    var books []Book
    rows, err := DB.Query("SELECT Book_id, Title, Summary, Author, First_published, Last_updated FROM books")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failing to query books from DB: " + err.Error()})
        return nil, fmt.Errorf("failing to query books from DB: %v", err)
    }
    defer rows.Close()
    for rows.Next() {
        var book Book
        if err := rows.Scan(&book.Book_id, &book.Title, &book.Summary, &book.Author, &book.First_published, &book.Last_updated); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failing to scan books: " + err.Error()})
            return nil, fmt.Errorf("failing to scan book: %v", err)
        }
        books = append(books, book)
    }
    return books, nil
}
func showBooksWithHTML(c *gin.Context) {
    books, err := loadBooksFromDB(c)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    tmpl, err := template.ParseFiles("templates/index.html")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Set Content-Type to text/html and execute the template with data
    c.Header("Content-Type", "text/html")

    if err := tmpl.Execute(c.Writer, books); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
}

func deleteBooks(c *gin.Context) {
    bookIDs := c.PostFormArray("book_ids")
    if len(bookIDs) == 0 {
        c.String(http.StatusBadRequest, "No books selected for deletion")
        return
    }
    var bookInts = make([]int,len(bookIDs))
    for i := 0; i < len(bookInts); i++ {
        bookInts[i], err = strconv.Atoi(bookIDs[i])
        if err != nil {
            c.String(http.StatusBadRequest, "Non-Intable book ID")
            return
        }
    }
    
    query := "DELETE FROM books WHERE Book_id IN (?)"
    query, args, err := sqlx.In(query, bookInts)
    if err != nil {
        c.String(http.StatusInternalServerError, err.Error())
        return
    }

    _, err = DB.Exec(query, args...)
    if err != nil {
        c.String(http.StatusInternalServerError, err.Error())
        return
    }

    c.Redirect(http.StatusSeeOther, "/books")
}

func main() {
    if err := InitDB(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer DB.Close()
    router := gin.Default()
    // Configure CORS
    config := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"}, // Replace with your Next.js frontend URL
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization"},
        AllowCredentials: true,
        MaxAge:           300, // Maximum age for browser to cache preflight requests
    })

    // Use CORS middleware
    router.Use(func(c *gin.Context) {
        config.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            c.Next()
        })).ServeHTTP(c.Writer, c.Request)
    })
    router.Static("/static", "./static")
    router.LoadHTMLGlob("templates/*")

    router.GET("/", showForm)
    router.GET("/books", showBooksWithHTML)
    router.POST("/books", CreateBookFromForm)
    router.POST("/api/books", CreateBook)
    router.GET("/api/books", getBooks)
    router.POST("/delete-multiple", deleteBooks)
    log.Println("Server starting on http://localhost:8080")
    if err := router.Run("localhost:8080"); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}