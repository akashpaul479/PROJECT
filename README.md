# College Management System (Go Backend Project)

This project is a College Management System REST API built using **Go (Golang)**, MySQL, Redis, and JWT Authentication.  

It helps manage:  
- Students  
- Lecturers  
- Library Books  
- Borrow & Return Records  
- User Login System

This project is designed to learn real backend development concepts.
***

# Why I Built This Project  
I created this project to learn:

- How REST APIs work in Go  
- How to connect MySQL with Go  
- How to use Redis for caching  
- How JWT authentication works  
- How to secure APIs  
- How to write Swagger documentation  
- How to structure a backend project
***

# Technologies Used  
| Tool        | Purpose               |
| ----------- | --------------------- |
| Go (Golang) | Backend Programming   |
| MySQL       | Store Data            |
| Redis       | Cache Data            |
| JWT         | Authentication        |
| Gorilla Mux | Routing               |
| Swagger     | API Docs              |
| Godotenv    | Environment Variables |

***

# Project Structure
college_management_system/

├── main.go                  -> Program Entry Point
│
├── college_management_system/
│   ├── server.go            -> Server Setup
│   ├── auth.go              -> Login & JWT
│   ├── student.go           -> Student APIs
│   ├── lecturer.go          -> Lecturer APIs
│   ├── library.go           -> Library APIs
│   ├── borrow.go            -> Borrow System
│   └── middleware.go        -> JWT Middleware
│
├── docs/                    -> Swagger Files
├── .env                     -> Environment Variables
└── README.md  

- This structure helps me keep each feature in a separate file. 
*** 
# Environment Variables (.env File)  
### Create a .env file:  
```.env
MYSQL_DSN=root:root@tcp(localhost:3306)/db_name
REDIS_ADDR=localhost:6379

JWT_SECRET=mysecretkey

EMAIL=admin@gmail.com
PASSWORD=admin123
```  
## why?
| Variable   | Use                 |
| ---------- | ------------------- |
| MYSQL_DSN  | Database Connection |
| REDIS_ADDR | Redis Server        |
| JWT_SECRET | Sign Tokens         |
| EMAIL      | Login User          |
| PASSWORD   | Login Password      |  
- Keeping secrets in .env is more secure.  
***
# How to Run My Project  
### Step 1: Clone  
```bash
git clone <your-repo-url>
cd college_management_system
```
### Step 2: Install Packages  
```bash
go mod tidy
```
### Step 3: Generate Swagger  
```bash
swag init
```
### Step 4: Run Server  
```bash
go run main.go
```
### Server runs on:  
```bash
http://localhost:8080
```
***

# Swagger API Docs
### Open in browser:  
```bash
http://localhost:8080/swagger/index.html
```
- Here I can test APIs without Postman.  
***
# Authentication System (JWT)  
### My project uses:   
- Access Token (15 min)  
- Refresh Token (7 days)  
- Stored in Cookies  

## Login API
```bash
POST /login
```
### Example:  
```bash
{
  "email": "admin@gmail.com",
  "password": "admin123"
}
```
After login → cookies are set → protected APIs work.  
***

# API Endpoints  
### Authentication  
| Method | URL      | Work      |
| ------ | -------- | --------- |
| POST   | /login   | Login     |
| POST   | /refresh | New Token |
| POST   | /logout  | Logout    |  

### Students  
| Method | URL                | Work        |
| ------ | ------------------ | ----------- |
| POST   | /api/students      | Add Student |
| GET    | /api/students      | View All    |
| GET    | /api/students/{id} | View One    |
| PUT    | /api/students/{id} | Update      |
| DELETE | /api/students/{id} | Delete      |  

### Lecturers  
| Method | URL                 | Work   |
| ------ | ------------------- | ------ |
| POST   | /api/lecturers      | Add    |
| GET    | /api/lecturers      | View   |
| PUT    | /api/lecturers/{id} | Update |
| DELETE | /api/lecturers/{id} | Delete |  

### Library  
| Method | URL                 | Work      |
| ------ | ------------------- | --------- |
| POST   | /api/libraries      | Add Book  |
| GET    | /api/libraries/{id} | View Book |
| PUT    | /api/libraries/{id} | Update    |
| DELETE | /api/libraries/{id} | Delete    |  

### Borrow System  
| Method | URL         | Work        |
| ------ | ----------- | ----------- |
| POST   | /api/borrow | Borrow Book |
| GET    | /api/borrow | History     |
| POST   | /api/return | Return Book |  
***

# Redis Caching  
I use Redis to cache:  
- Students  
- Lecturers  
- Library  
### Flow:  
```bash
Client → Redis → MySQL → Redis → Client
```
TTL(Time to live) = 10 seconds  
This improves speed.  
***

# Logging & Audit  
#### My project logs actions:  
### Example:  
```bash
[LOG] CREATE_STUDENT
[AUDIT] CREATE STUDENT 5
```
This helps in debugging.  
***

# Testing with Curl  
## JWT Authentication And Authorization
### Login  
```bash
curl -X POST -H "Content-Type: application/json" ^
-d "{\"email\":\"admin@gmail.com\",\"password\":\"admin123\"}" ^
 http://localhost:8080/login -c cookies.txt
```
### Refresh 
```bash 
curl -X POST http://localhost:8080/refresh -b cookies.txt
```
### Logout  
```bash
curl -X POST http://localhost:8080/logout -b cookies.txt
```
  ## Students CRUD Operations
### Create Students  
```bash
curl -X POST -H "Content-Type: application/json" ^
 -d "{\"name\":\"example\",\"age\":45,\"email\":\"example@gmail.com\",\"dept\":\"CSE\"}" ^  
 http://localhost:8080/api/students -b cookies.txt  
```
### Get all Students  
```bash
curl http://localhost:8080/api/students -b cookies.txt
```
### Get students by ID  
```bash
curl http://localhost:8080/api/students/1 -b cookies.txt  
```
### Update Students  
```bash
curl -X PUT -H "Content-Type: application/json" ^
-d "{\"name\":\"john\",\"age\":50,\"email\":\"john@gmail.com\",\"dept\":\"ECE\"}" ^
http://localhost:8080/api/students/1 -b cookies.txt
```
### Delete Students  
```bash
curl -X DELETE http://localhost:8080/api/students/1 -b cookies.txt
```

## Lecturers CRUD Operations
### Create Lecturers 
```bash
curl -X POST -H "Content-Type: application/json" ^
 -d "{\"name\":\"example\",\"age\":45,\"email\":\"example@gmail.com\",\"designation\":\"Professor\"}" ^  
 http://localhost:8080/api/lecturers -b cookies.txt  
```
### Get all Lecturers  
```bash
curl http://localhost:8080/api/lecturers  -b cookies.txt
```
### Get Lecturers by ID  
```bash
curl http://localhost:8080/api/lecturers/1 -b cookies.txt  
```
### Update Lecturers 
```bash
curl -X PUT -H "Content-Type: application/json" ^
-d "{\"name\":\"john\",\"age\":50,\"email\":\"john@gmail.com\",\"designation\":\"HOD\"}" ^
http://localhost:8080/api/lecturers/1 -b cookies.txt
```
### Delete Lecturers  
```bash
curl -X DELETE http://localhost:8080/api/lecturers/1 -b cookies.txt
```

## Library CRUD Operations
### Create Library
```bash
curl -X POST -H "Content-Type: application/json" ^
-d "{\"book_name\":\"The Guide\",\"title\":\"tourist guide\",\"author\":\" R.K. Narayan\",\"available_copies\":10}" ^
http://localhost:8080/api/libraries -b cookies.txt
```
### Get Library by ID  
```bash
curl http://localhost:8080/api/libraries/1 -b cookies.txt  
```
### Update Library
```bash
curl -X PUT -H "Content-Type: application/json" ^
-d "{\"book_name\":\"The boys\",\"title\":\"boys\",\"author\":\" john\",\"available_copies\":15}" ^
http://localhost:8080/api/libraries/1 -b cookies.txt
```
### Delete Library 
```bash
curl -X DELETE http://localhost:8080/api/libraries/1 -b cookies.txt
```
## Borrow_Records
```bash
curl -X POST -H "Content-Type: application/json" ^
-d "{\"user_id\":1,\"user_type\":\"student\",\"book_id\":1}" ^
http://localhost:8080/api/borrow -b cookies.txt
```
## Get all Borrow_Record  
```bash
curl -X GET http://localhost:8080/api/borrow -b cookies.txt
```
## Return_Records  
```bash
curl -X POST -H "Content-Type: application/json" ^
-d "{\"user_id\":1,\"user_type\":\"student\",\"book_id\":1}" ^
http://localhost:8080/api/return -b cookies.txt
```
***
## Status Code   
| Range | Meaning         | Example     |
| ----- | --------------- | ----------- |
| 1xx   | Info            | Rare        |
| 2xx   | Success         | 200, 201    |
| 3xx   | Redirect        | Rare in API |
| 4xx   | Client Error    | 400, 401    |
| 5xx   | Server Error    | 500         |  
***
# Contributions  
Contributions are Welcome!  
- Fork the repository  
- Create a future branch  
- Commit changes  
- Push and open a pull Request  
***
# License  
This project is licensed under MIT License.  
***
# Thanks  
Thank you for visiting the College_Management_System repository. Feel free to reach out it you want help setting up, extending, or deploying this project!













