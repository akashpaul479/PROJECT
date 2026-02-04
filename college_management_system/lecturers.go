package collegemanagementsystem

import (
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

// Lecturer represents a lecturer entity stored in MySQL
type Lecturer struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Age         int    `json:"age"`
	Email       string `json:"email"`
	Designation string `json:"designation"`
}

// validationLecturer validates incoming lecturer data
func Validatelecturer(lecturer Lecturer) error {

	// validate name
	if strings.TrimSpace(lecturer.Name) == "" {
		return fmt.Errorf("name is invalid and empty")
	}
	// validate email
	if lecturer.Email == "" {
		return fmt.Errorf("email is invalid and empty")
	}
	if !strings.HasSuffix(lecturer.Email, "@gmail.com") {
		return fmt.Errorf("email is invalid and does not contain @gmail.com")
	}
	prefix := strings.TrimSuffix(lecturer.Email, "@gmail.com")
	if prefix == "" {
		return fmt.Errorf("email must contains a prefix before the @gmail.com ")
	}
	// validate age
	if lecturer.Age <= 0 {
		return fmt.Errorf("Invalid age , age is less than 0")
	}
	if lecturer.Age >= 100 {
		return fmt.Errorf("Invalid age , age is grater than 100")
	}
	// validate designation
	if lecturer.Designation == "" {
		return fmt.Errorf("Year is invalid, please enter a valid year")
	}
	return nil
}

// createlecturerHandler handles ctreation of a new lecturer
func (h *HybridHandler) CreateLecturerHandler(w http.ResponseWriter, r *http.Request) {

	// Decode incoming JSON requests body
	var lecturers Lecturer
	if err := json.NewDecoder(r.Body).Decode(&lecturers); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate Requests payload
	if err := Validatelecturer(lecturers); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
		return
	}

	// Insert lecturer records into MySQL database
	res, err := h.MySQL.db.Exec("INSERT INTO lecturers (name , email , age , designation) VALUES (? , ? , ? , ? )", lecturers.Name, lecturers.Email, lecturers.Age, lecturers.Designation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// auto generated ID
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	lecturers.ID = int(id)

	// Log activity and Audit trail
	go LogActivity("CREATE_LECTURER", "system")
	go AuditLog("CREATE", "LECTURER", lecturers.ID, "system")

	// Send succes response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lecturers)
}

// GetLecturerHandler to get all students
func (a *HybridHandler) GetLecturerHandler(w http.ResponseWriter, r *http.Request) {

	// Execute query to fetch lecturers record
	rows, err := a.MySQL.db.Query("SELECT id , name , age , email , designation FROM lecturers")
	if err != nil {
		http.Error(w, "unable to fetch lecturers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var lecturers []Lecturer
	for rows.Next() {
		var L Lecturer
		if err := rows.Scan(&L.ID, &L.Name, &L.Age, &L.Email, &L.Designation); err != nil {
			http.Error(w, "rows scan failed", http.StatusInternalServerError)
			return
		}
		lecturers = append(lecturers, L)

	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lecturers)
}

// GetLecturerByIDHandler retrives a lecturer by id
func (h *HybridHandler) GetLecturerByIDHandler(w http.ResponseWriter, r *http.Request) {

	// Extrect lecturer id by URL
	vars := mux.Vars(r)
	id := vars["id"]

	// LogActivity
	go LogActivity("GET_LECTURER", "system")

	// Attempt to fetch from redis cache first
	value, err := h.Redis.Client.Get(h.Ctx, id).Result()
	if err == nil {
		log.Println("Cache hit!")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(value))
		return
	}

	// cache miss fetching from MySQL database
	fmt.Println("Cache miss Quering MySQL ...")
	row := h.MySQL.db.QueryRow("SELECT id , name , age , email  , designation FROM lecturers WHERE  id=?", id)

	var lecturers Lecturer
	if err := row.Scan(&lecturers.ID, &lecturers.Name, &lecturers.Age, &lecturers.Email, &lecturers.Designation); err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// marshal lecturer data for caching
	jsondata, err := json.Marshal(lecturers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// store result in redis cache (TTL)
	h.Redis.Client.Set(h.Ctx, id, jsondata, 10*time.Second)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsondata)
}

// updateLecturerHandler updates a exsisting lecturer
func (h *HybridHandler) UpdateLecturerHandler(w http.ResponseWriter, r *http.Request) {

	// Decode request body
	var lecturers Lecturer
	if err := json.NewDecoder(r.Body).Decode(&lecturers); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// validate updated data
	if err := Validatelecturer(lecturers); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
		return
	}

	// Execute Updated query
	res, err := h.MySQL.db.Exec("UPDATE lecturers SET name=?,email=?,age=?,designation=? WHERE id=?", lecturers.Name, lecturers.Email, lecturers.Age, lecturers.Designation, lecturers.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if records exists
	rows, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Update redis cache
	jsonData, err := json.Marshal(lecturers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	go h.Redis.Client.Set(h.Ctx, fmt.Sprint(lecturers.ID), jsonData, 10*time.Minute)

	// Log Update actions
	go LogActivity("UPDATE_LECTURER", "system")
	go AuditLog("UPDATE", "LECTURER", lecturers.ID, "system")

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// DeleteLecturerHandler deletes a lecturer by ID
func (h *HybridHandler) DeleteLecturerHandler(w http.ResponseWriter, r *http.Request) {

	// Extract id from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// convert id to integer
	idInt, _ := strconv.Atoi(id)

	// Execute delete query
	res, err := h.MySQL.db.Exec("DELETE FROM lecturers WHERE id=?", idInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check id lecturer exists
	rows, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "lecturer not found", http.StatusNotFound)
		return
	}

	// remove cache entry
	h.Redis.Client.Del(h.Ctx, id)

	// Log delete response
	go LogActivity("DELETE_LECTURER", "system")
	go AuditLog("DELETE", "LECTURER", idInt, "system")

	// send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("lecturer deleted"))
}
