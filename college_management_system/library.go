package collegemanagementsystem

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// library represents a library entity
type Library struct {
	Book_id          int    `json:"book_id"`
	Book_name        string `json:"book_name"`
	Title            string `json:"title"`
	Author           string `json:"author"`
	Available_copies int    `json:"available_copies"`
}

// borrow_records represents a borrowing transaction
type Borrow_records struct {
	Borrow_id   int          `json:"borrow_id"`
	User_id     int          `json:"user_id"`
	User_type   string       `json:"user_type"`
	Book_id     int          `json:"book_id"`
	Borrow_date sql.NullTime `json:"borrow_date"`
	Return_date sql.NullTime `json:"return_date"`
}

// validate library ensures that library input data is valid before DB operations
func ValidateLibrary(library Library) error {

	// validate book_name
	if strings.TrimSpace(library.Book_name) == "" {
		return fmt.Errorf("book_name is invalid and empty")
	}
	// validate title
	if strings.TrimSpace(library.Title) == "" {
		return fmt.Errorf("title is invalid and empty")
	}
	// validate author
	if strings.TrimSpace(library.Author) == "" {
		return fmt.Errorf("Author is invalid and empty")
	}
	// validate available_copies
	if library.Available_copies <= 0 {
		return fmt.Errorf("available_copies is less than 0")
	}
	return nil
}

// validateBorrowRecords ensures borrow record input is valid
func ValidateBorrowRecords(BR Borrow_records) error {
	// validate book_id
	if BR.Book_id <= 0 {
		return fmt.Errorf("invalid book_id")
	}
	// valkidate user_id
	if BR.User_id <= 0 {
		return fmt.Errorf("invalid user_id")
	}
	// validate user_type
	if BR.User_type == "" {
		return fmt.Errorf("user type cannot be empty")
	}
	// validate return_date
	if BR.Return_date.Valid && BR.Borrow_date.Valid {
		if BR.Return_date.Time.Before(BR.Borrow_date.Time) {
			return fmt.Errorf("return_date cannot be before borrow_date")
		}
	}
	return nil
}

// createLibraryHandler handles creation of a new library
func (h *HybridHandler) CreateLibraryHandler(w http.ResponseWriter, r *http.Request) {
	var libraries Library
	if err := json.NewDecoder(r.Body).Decode(&libraries); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidateLibrary(libraries); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
		return
	}
	res, err := h.MySQL.db.Exec("INSERT INTO libraries (book_name , title , author , available_copies) VALUES ( ? , ? , ? , ?)", libraries.Book_name, libraries.Title, libraries.Author, libraries.Available_copies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	libraries.Book_id = int(id)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(libraries)
}

// GetLibraryHandler retrives a library by id
func (h *HybridHandler) GetLibraryByIDHandler(w http.ResponseWriter, r *http.Request) {

	// Extract library id from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Log activity
	go LogActivity("GET_LIBRARY", "system")

	// attempt to fetch from redis cache first
	value, err := h.Redis.Client.Get(h.Ctx, id).Result()
	if err == nil {
		log.Println("Cache hit!")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(value))
		return
	}

	// cache miss querying MYSQL database
	fmt.Println("Cache miss Quering MySQL ...")
	row := h.MySQL.db.QueryRow("SELECT book_id ,book_name ,  title , author , available_copies FROM libraries WHERE  book_id=?", id)

	var libraries Library
	if err := row.Scan(&libraries.Book_id, &libraries.Book_name, &libraries.Title, &libraries.Author, &libraries.Available_copies); err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	//  marshal library data for caching
	jsondata, err := json.Marshal(libraries)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	go h.Redis.Client.Set(h.Ctx, id, jsondata, 10*time.Second)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsondata)
}

// BorrowrecordsHandler handles
func (h *HybridHandler) BorrowRecordsHandler(w http.ResponseWriter, r *http.Request) {

	// Decode requests body
	var record Borrow_records
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	// validate user type
	if record.User_type != "student" && record.User_type != "lecturer" {
		http.Error(w, "invalid user_type, must be 'student' or 'lecturer'", http.StatusBadRequest)
		return
	}

	// validate borrow record
	if err := ValidateBorrowRecords(record); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//  Check if Book is available
	var available int
	err := h.MySQL.db.QueryRow("SELECT available_copies FROM libraries WHERE book_id=?", record.Book_id).Scan(&available)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}
	if available <= 0 {
		http.Error(w, " Book not available", http.StatusBadRequest)
		return
	}
	// Insert borrow record
	_, err = h.MySQL.db.Exec("INSERT INTO borrow_records(user_id, user_type,book_id ,borrow_date)VALUES (? , ? , ? , NOW())", record.User_id, record.User_type, record.Book_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//  decrement available copies
	_, err = h.MySQL.db.Exec("UPDATE libraries SET available_copies = available_copies-1 WHERE book_id=?", record.Book_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log Activity and audit trails
	go LogActivity("BORROW_RECORD", "system")
	go AuditLog("BORROW", "RECORDS", record.Book_id, "system")

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "Book borrowed!"})
}

// Get all borrowrecords retrives all borrow history from the database
func (h *HybridHandler) GetBorrowRecordsHandler(w http.ResponseWriter, r *http.Request) {

	// Execute SQL query to fetch borrow records
	rows, err := h.MySQL.db.Query("SELECT b.borrow_id, b.user_id, b.user_type, b.book_id, l.book_name, b.borrow_date, b.return_date FROM borrow_records b JOIN libraries l ON b.book_id=l.book_id ORDER BY b.borrow_id DESC")

	// if query fails , return server error
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// close rows when functiions ends
	defer rows.Close()

	// Create struct to store one borrow record
	type borrowInfo struct {
		BorrowID   int          `json:"borrow_id"`
		UserID     int          `json:"user_id"`
		UserType   string       `json:"user_type"`
		BookID     int          `json:"book_id"`
		BookType   string       `json:"book_type"`
		BorrowDate sql.NullTime `json:"borrow_date"`
		ReturnDate sql.NullTime `json:"return_date"`
	}

	// slice to store multiple borrow records
	var records []borrowInfo

	// loop through all database rows
	for rows.Next() {

		var r borrowInfo

		// read column values into struct feilds
		err := rows.Scan(&r.BorrowID, &r.UserID, &r.UserType, &r.BookID, &r.BookType, &r.BorrowDate, &r.ReturnDate)
		// if scanning fails return error
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// add records to slice
		records = append(records, r)
	}

	// send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

// Return book
func (h *HybridHandler) ReturnRecordsHandler(w http.ResponseWriter, r *http.Request) {

	// Decode requests body
	var record Borrow_records
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, "invalid json", http.StatusInternalServerError)
		return
	}
	// validate user type
	if record.User_type != "student" && record.User_type != "lecturer" {
		http.Error(w, "invalid user_type, must be 'student' or 'lecturer'", http.StatusBadRequest)
		return
	}
	// Update borrow_books record with return date
	res, err := h.MySQL.db.Exec("UPDATE borrow_records SET return_date=CURDATE() WHERE user_id=? AND book_id=? AND return_date IS NULL", record.User_id, record.Book_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		http.Error(w, "no active borrow record found", http.StatusNotFound)
		return
	}
	//  increment available copies
	_, err = h.MySQL.db.Exec("UPDATE libraries SET available_copies = available_copies+1 WHERE book_id=?", record.Book_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// update redis cache
	jsonData, err := json.Marshal(record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	go h.Redis.Client.Set(h.Ctx, fmt.Sprint(record.Book_id), jsonData, 10*time.Second)

	// Log Activity and audit trails
	go LogActivity("RETURN_RECORD", "system")
	go AuditLog("RETURN", "RECORDS", record.Book_id, "system")

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "Book returned!"})

}
