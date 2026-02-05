package collegemanagementsystem

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	Borrow_id   int    `json:"borrow_id"`
	User_id     int    `json:"user_id"`
	User_type   string `json:"user_type"`
	Book_id     int    `json:"book_id"`
	Borrow_date string `json:"borrow_date"`
	Return_date string `json:"return_date"`
}

// Create struct to store one borrow record
type BorrowInfo struct {
	BorrowID   int    `json:"borrow_id"`
	UserID     int    `json:"user_id"`
	UserType   string `json:"user_type"`
	BookID     int    `json:"book_id"`
	BookType   string `json:"book_type"`
	BorrowDate string `json:"borrow_date"`
	ReturnDate string `json:"return_date"`
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
	return nil
}

// CreateLibraryHandler godoc
// @Summary Add library book
// @Tags Library
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param library body Library true "Library Book"
// @Success 201 {object} Library
// @Router /api/libraries [post]
// createLibraryHandler handles creation of a new library
func (h *HybridHandler) CreateLibraryHandler(w http.ResponseWriter, r *http.Request) {

	// Decode incoming JSON requests body
	var libraries Library
	if err := json.NewDecoder(r.Body).Decode(&libraries); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// validate requests payload
	if err := ValidateLibrary(libraries); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
		return
	}

	// Insert library records into MySQL database
	res, err := h.MySQL.db.Exec("INSERT INTO libraries (book_name , title , author , available_copies) VALUES ( ? , ? , ? , ?)", libraries.Book_name, libraries.Title, libraries.Author, libraries.Available_copies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Auto generated id
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

// GetLibraryByIDHandler godoc
// @Summary Get library by ID
// @Tags Libraries
// @Security BearerAuth
// @Produce json
// @Param id path int true "Library ID"
// @Success 200 {object} Library
// @Failure 404 {object} map[string]string
// @Router /api/Libraries/{id} [get]
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

// UpdateLibraryHandler godoc
// @Summary Update library
// @Tags Libraries
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param library body Library true "Updated Library"
// @Success 200 {object} Library
// @Router /api/Libraries/{id} [put]
// UpdateLibraryHandler updates an existing library record by ID
func (h *HybridHandler) UpdateLibraryHandler(w http.ResponseWriter, r *http.Request) {

	// Extract library id from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Decode incoming JSON requests body
	var libraries Library
	if err := json.NewDecoder(r.Body).Decode(&libraries); err != nil {
		http.Error(w, "failed to decode response", http.StatusInternalServerError)
		return
	}

	// validate updated data
	if err := ValidateLibrary(libraries); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// update MySQL record
	_, err := h.MySQL.db.Exec("UPDATE libraries SET book_name=? , title=? , author=? , available_copies=? WHERE book_id=?", libraries.Book_name, libraries.Title, libraries.Author, libraries.Available_copies, id)
	if err != nil {
		http.Error(w, "failed to update", http.StatusInternalServerError)
		return
	}
	libraries.Book_id, _ = strconv.Atoi(id)

	// refresh redis cache
	jsonData, err := json.Marshal(libraries)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log update actions
	go LogActivity("UPDATE_STUDENT", "system")
	go AuditLog("UPDATE", "STUDENT", libraries.Book_id, "system")

	go h.Redis.Client.Set(h.Ctx, id, jsonData, 10*time.Second)

	// send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "library updated succesfully"})
}

// DeleteLibraryHandler godoc
// @Summary Delete library
// @Tags Libraries
// @Security BearerAuth
// @Produce json
// @Param id path int true "Library ID"
// @Success 200 {string} string "library deleted"
// @Router /api/libraries/{id} [delete]
// DeleteLibraryHandler dleted a library record by ID
func (h *HybridHandler) DeleteLibraryHandler(w http.ResponseWriter, r *http.Request) {

	// Extract library id from URl
	vars := mux.Vars(r)
	id := vars["id"]

	IdInt, _ := strconv.Atoi(id)
	// delete borrow_records
	_, err := h.MySQL.db.Exec("DELETE FROM borrow_records WHERE book_id=?", IdInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete library
	res, err := h.MySQL.db.Exec("DELETE FROM libraries WHERE book_id=?", IdInt)
	if err != nil {
		http.Error(w, "unable to delete", http.StatusInternalServerError)
		return
	}

	// check if library exsists
	rows, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if rows == 0 {
		http.Error(w, "library not found", http.StatusNotFound)
		return
	}

	// Invalid redis cache
	go h.Redis.Client.Del(h.Ctx, id)

	// Log delete response
	go LogActivity("DELETE_STUDENTS", "system")
	go AuditLog("DELETE", "STUDENT", IdInt, "system")

	// send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "library daleted successfully!"})

}

// BorrowRecordsHandler godoc
// @Summary Borrow book
// @Tags Borrow
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param record body Borrow_records true "Borrow Record"
// @Success 201 {object} map[string]string
// @Router /api/borrow [post]
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

// GetBorrowRecordsHandler godoc
// @Summary Get all borrow records
// @Description Retrieve complete borrowing history with book details
// @Tags Borrow
// @Security BearerAuth
// @Produce json
// @Success 200 {array} BorrowInfo
// @Failure 500 {object} map[string]string
// @Router /api/borrow [get]
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

	// slice to store multiple borrow records
	var records []BorrowInfo

	// loop through all database rows
	for rows.Next() {

		var r BorrowInfo
		var borrowdate, returndate sql.NullTime

		// read column values into struct feilds
		err := rows.Scan(&r.BorrowID, &r.UserID, &r.UserType, &r.BookID, &r.BookType, &borrowdate, &returndate)
		// if scanning fails return error
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if borrowdate.Valid {
			r.BorrowDate = borrowdate.Time.Format(time.RFC3339)
		}

		if returndate.Valid {
			r.ReturnDate = returndate.Time.Format(time.RFC3339)
		}

		// add records to slice
		records = append(records, r)
	}

	// send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

// ReturnRecordsHandler godoc
// @Summary Return book
// @Tags Borrow
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param record body Borrow_records true "Return Record"
// @Success 201 {object} map[string]string
// @Router /api/return [post]
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
