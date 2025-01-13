package main
import (
    "fmt"
    _ "github.com/go-sql-driver/mysql"
    "net/http"
	"github.com/gin-gonic/gin"
    "time"
    "log"
    //"net/http/httptest"
    //"encoding/json"
    //"strings"
    "github.com/jmoiron/sqlx"
    "html/template"
    "net/url"
    "strconv"
)


func helloHandler(w http.ResponseWriter, r *http.Request, extraMessage string){
    fmt.Fprintf(w, "Hello, HTTP World!\n\tThe time is %s\n\tMessage: %s", time.Now(), extraMessage)
}


/*func (cd *CustomDate) UnmarshalJSON(b []byte) error {
    fmt.Println("UnmarshallJson: " + string(b))
    log.Printf("UnmarshallJSON: %s", string(b))

    // Trim quotes from JSON string
    dateStr := string(b[1 : len(b)-1])
    
    // Parse the date string
    t, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        return err
    }
    
    // Set the time to UTC
    cd.Time = t.UTC()
    return nil
}*/

type CustomTime struct {
    time.Time
}

func (ct *CustomTime) Unmarshal(b []byte) error {
    fmt.Println(string(b))
    if string(b) == "null" {
        ct.Time = time.Time{}
        return nil
    }
    
    t, err := time.Parse("2006-01-02", string(b))
    if err != nil {
        return err
    }
    ct.Time = t
    return nil
}


type Book struct {
    Book_id   int64 `json:"Book_id"`
    Title string `json:"Title"`
    Summary string `json:"Summary"`
    Author string`json:"Author"`
    First_published time.Time `json:"First_published,omitempty"`
    Last_updated time.Time `json:"Last_updated,omitempty"`
}

func (b Book) String() string {
    return fmt.Sprintf("Title: %s, Summary: %s, Author: %s, First_published: %s, Last_update: %s", b.Title, b.Summary, b.Author, b.First_published, b.Last_updated)
}

var DB *sqlx.DB
var err error
var books []Book

func getBooks(c *gin.Context) {
	c.IndentedJSON(http.StatusOK,books)
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

func showBooks(c *gin.Context) {

    // Query the database for books
    rows, err := DB.Query("SELECT Book_id, Title, Summary, Author, Last_updated, First_published FROM books")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var books []Book
    for rows.Next() {
        var book Book
        if err := rows.Scan(&book.Book_id, &book.Title, &book.Summary, &book.Author, &book.First_published, &book.Last_updated); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        books = append(books, book)
    }

    // Parse and execute the template
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

func deleteBook(c *gin.Context) {
    id := c.Param("id")
    idInt, _ := strconv.Atoi(id)
    _, err = DB.Exec("DELETE FROM books WHERE Book_id = ?", idInt)
    if err != nil {
        c.String(http.StatusInternalServerError, err.Error())
        return
    }
    showBooks(c)
    c.Redirect(http.StatusSeeOther, "/books")
}

func main() {
    if err := InitDB(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer DB.Close()
    router := gin.Default()
	//router.POST("/books", postBookTest)
    router.GET("/", showForm)
    router.POST("/books", CreateBookFromForm)
    router.GET("/books", showBooks)
    router.POST("/delete/:Book_id", deleteBook)
    router.Run("localhost:8080")
}
/*
func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/book", func(c *gin.Context) {
		c.String(200, "testbook")
	})
	return r
}
func postBookTest(c *gin.Context) {
    router := setupRouter()
    router.POST("/books",postBook)

    w := httptest.NewRecorder()

    exampleBook := Book{
        Title: "The Catcher in the Rye",
        Author: "J.D. Salinger", 
        First_published: "1951",
        Summary: "The story of Holden Caulfield, a teenage boy who struggles with alienation and loss of innocence.",
        Last_updated: time.Now().Format("2006-01-02 15:04:05"),
    }

    bookJson,_ := json.Marshal(exampleBook)
    req, _ := http.NewRequest("POST", "/books", strings.NewReader(string(bookJson)))
    router.ServeHTTP(w, req)
	var actualResponse struct{
		Data Book `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &actualResponse)
    //assert.NoError(t, err, "Failed to unmarshal actual response")
    //assert.Equal(c, 200, w.Code)
    // Compare the book data
	postBookToDB(exampleBook)
	//assert.Equal(c, exampleBook, actualResponse.Data, "Book data doesn't match")

func postBook(c *gin.Context) {
    var book Book
    if err := c.ShouldBindJSON(&book); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": book})
    postBookToDB(book)
}

func postBookToDB(book Book) {
    DB, err := sql.Open("mysql","root:bricha101@tcp(127.0.0.1:3306)/library_system")
    if err != nil {
        // Handle error
    }
    defer DB.Close()

    stmt, err := DB.Prepare("INSERT INTO Books (Title, Summary, Author, First_published, Last_updated) VALUES (?, ?, ?, ?, ?)")
    if err != nil {
        // Handle error
    }
    defer stmt.Close()

    stmt.Exec(book.Title, book.Summary, book.Author, book.First_published, book.Last_updated)
    return
}
func initDB() {
    dsn := "root:bricha101@tcp(127.0.0.1:3306)/library_system"
    DB, err := sql.Open("mysql", dsn)
    if err != nil {
        panic(err.Error())
    }
    defer DB.Close()

    err = DB.Ping()
    if err != nil {
        panic(err.Error())
    }
    
    // Select data from the table
    rows, err := DB.Query("SELECT * FROM books")
    if err != nil {
        panic(err.Error())
    }
    row := DB.QueryRow("SELECT * FROM books WHERE book_id = ?", 1)
    var book Book
    err = row.Scan(&book.Book_id, &book.Title, &book.Summary, &book.Author, &book.First_published, &book.Last_updated)
    //print("the book is:")
    //fmt.Println(book)
    books = append(books,book)
    //print("the collection \"Books\" is:")
    //fmt.Println(books)
    if err != nil {
        if err == sql.ErrNoRows {
            fmt.Println("No rows found")
        } else {
            panic(err.Error())
        }
    }
    //fmt.Sprintf("ID: %d, Title: %s, Summary: %s, Author: %s, First Published: %s, Last Updated: %s\n", book.book_id, book.title, book.summary, book.author, book.first_published, book.last_updated)

    err = rows.Err()
    if err != nil {
        panic(err.Error())
    }
    
}
*/