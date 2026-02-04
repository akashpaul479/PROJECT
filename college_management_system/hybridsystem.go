package collegemanagementsystem

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// MySQLInstance  wraps a MySQL database connections
type MySQLInstance struct {
	db *sql.DB
}

// Redis Instance wraps a redis client.
type RedisInstance struct {
	Client *redis.Client
}

// HybridHandler aggregates MySQL , MongoDB , Redis instances along with a shared context.
type HybridHandler struct {
	MySQL *MySQLInstance
	Redis *RedisInstance
	Ctx   context.Context
}

// connectMySQL initilizes a MySQL connection using DSN from environment variables.
func ConnectMySQL() (*MySQLInstance, error) {
	db, err := sql.Open("mysql", os.Getenv("MYSQL_DSN"))
	if err != nil {
		log.Panic(err)
	}
	return &MySQLInstance{db: db}, nil
}

// ConnectRedis initilizes a Redis Client environment variables
func ConnectRedis() (*RedisInstance, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
		DB:   0,
	})
	return &RedisInstance{Client: rdb}, nil
}

// Background Utilities
// Logging (gouroutine safe)
func LogActivity(action, actor string) {
	log.Printf("[LOG] %s by %s at %s\n", action, actor, time.Now())
}

// Audit trail
func AuditLog(action, entity string, id any, actor string) {
	log.Printf("[AUDIT] action=%s entity=%s id=%v actor=%s time=%s\n", action, entity, id, actor, time.Now())
}

// main function
func CollegeManagementSystem() {

	// Load environment variables from .env file
	godotenv.Load()

	// Ensures JWT Secret  is set
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is not set or empty")
	}
	SecretKey = []byte(secret)

	// Initilizes Redis
	redisinstance, err := ConnectRedis()
	if err != nil {
		panic(err)
	}

	// Initilizes MySQL
	mysqlinstance, err := ConnectMySQL()
	if err != nil {
		panic(err)
	}

	// Create handler with all DB instanmces
	handler := &HybridHandler{Redis: redisinstance, MySQL: mysqlinstance, Ctx: context.Background()}

	// Setup HTTP routers
	r := mux.NewRouter()

	// Authentication routes
	r.HandleFunc("/login", LoginHandler).Methods("POST")
	r.HandleFunc("/refresh", RefreshHandler).Methods("POST")
	r.HandleFunc("/logout", LogoutHandler).Methods("POST")

	// Protected route
	api := r.PathPrefix("/api").Subrouter()
	api.Use(JwtMiddleware)

	// Student CRUD routes
	api.HandleFunc("/students", handler.CreateStudentHandler).Methods("POST")
	api.HandleFunc("/students", handler.GetStudentHandler).Methods("GET")
	api.HandleFunc("/students/{id}", handler.GetstudentByIDHandler).Methods("GET")
	api.HandleFunc("/students/{id}", handler.UpdateStudentHandler).Methods("PUT")
	api.HandleFunc("/students/{id}", handler.DeleteStudentHandler).Methods("DELETE")

	// Lecturer CRUD routes
	api.HandleFunc("/lecturers", handler.CreateLecturerHandler).Methods("POST")
	api.HandleFunc("/lecturers", handler.GetLecturerHandler).Methods("GET")
	api.HandleFunc("/lecturers/{id}", handler.GetLecturerByIDHandler).Methods("GET")
	api.HandleFunc("/lecturers/{id}", handler.UpdateLecturerHandler).Methods("PUT")
	api.HandleFunc("/lecturers/{id}", handler.DeleteLecturerHandler).Methods("DELETE")

	// Library routes
	api.HandleFunc("/libraries", handler.CreateLibraryHandler).Methods("POST")
	api.HandleFunc("/libraries/{id}", handler.GetLibraryByIDHandler).Methods("GEt")

	// Borrow_records routes
	api.HandleFunc("/borrow", handler.BorrowRecordsHandler).Methods("POST")
	api.HandleFunc("/borrow", handler.GetBorrowRecordsHandler).Methods("GET")
	api.HandleFunc("/return", handler.ReturnRecordsHandler).Methods("POST")

	fmt.Println("Server running on port:8080")
	http.ListenAndServe(":8080", r)
}
