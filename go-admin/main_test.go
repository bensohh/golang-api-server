package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bensohh/go-admin/controllers"
	"github.com/bensohh/go-admin/models"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func setup() {
	godotenv.Load()

	models.ConnectDatabase()
	insertTestData()
}
func createAndLoad() {
	teardown()
	models.DB.AutoMigrate(&models.Teacher{})
	models.DB.AutoMigrate(&models.Student{})
	models.DB.AutoMigrate(&models.Registry{})
	insertTestData()
}

func teardown() {
	models.DB.Migrator().DropTable(&models.Registry{})
	models.DB.Migrator().DropTable(&models.Student{})
	models.DB.Migrator().DropTable(&models.Teacher{})
}

func insertTestData() {
	teachers := []models.Teacher{
		{Name: "Ken", Email: "teacherken@gmail.com"},
		{Name: "Joe", Email: "teacherjoe@gmail.com"},
	}
	models.DB.Create(&teachers)

	students := []models.Student{
		{Name: "Jon", Email: "studentjon@gmail.com"},
		{Name: "Hon", Email: "studenthon@gmail.com"},
		{Name: "Tom", Email: "studenttom@gmail.com"},
		{Name: "Stu1", Email: "studentunderkenonly@gmail.com"},
	}
	models.DB.Create(&students)

	registries := []models.Registry{
		{TeacherEmail: "teacherjoe@gmail.com", StudentEmail: "studentjon@gmail.com"},
		{TeacherEmail: "teacherjoe@gmail.com", StudentEmail: "studenthon@gmail.com"},
	}
	models.DB.Create(&registries)
}

func TestMain(m *testing.M) {
	setup()

	code := m.Run()

	teardown()

	os.Exit(code)
}

// Tests if the server alive response is correct
func TestServerAlive(t *testing.T) {
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/", controllers.TestServer).Methods("GET")
	router.ServeHTTP(response, request)

	assert.Equal(t, 200, response.Code, "OK response is expected")
	assert.Equal(t, "Server is running", strings.Trim(response.Body.String(), "\"\n"), "Expected server is running")
}

func TestCheckStudentExists(t *testing.T) {
	// Set-up Test Data
	createAndLoad()

	// Case when: Student exists
	exists := controllers.CheckStudentExists("studentjon@gmail.com")
	assert.True(t, exists, "Expect student to exist")
}

func TestCheckStudentDontExists(t *testing.T) {
	// Set-up Test Data
	createAndLoad()

	// Case when: Student does not exists
	fmt.Println("Logs a record not found below (correct behaviour)")
	notExists := controllers.CheckStudentExists("nonexistentstudent@gmail.com")
	assert.False(t, notExists, "Expect student to not exist")
}

func TestRegisterStudents(t *testing.T) {
	requestBody := controllers.RegisterStudentsRequest{
		Teacher:  "teacherken@gmail.com",
		Students: []string{"studentjon@gmail.com", "studenthon@gmail.com", "studentunderkenonly@gmail.com"},
	}
	// Set-up Test Data
	createAndLoad()

	// Assert that initial registry does not contain data (Should log down the error log from no data found)
	var initRegistry models.Registry
	fmt.Println("Logs a record not found below (correct behaviour)")
	err := models.DB.Where("teacher_email = ?", requestBody.Teacher).First(&initRegistry).Error
	assert.Error(t, err, "Expected error when no data found")

	jsonStr, _ := json.Marshal(requestBody)
	request, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonStr))
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/register", controllers.RegisterStudents).Methods("POST")
	router.ServeHTTP(response, request)
	assert.Equal(t, 204, response.Code, "OK response is expected")

	// Assert that the registry data is added to the database
	var registry models.Registry
	models.DB.Where("teacher_email = ?", requestBody.Teacher).First(&registry)
	assert.Contains(t, requestBody.Students, registry.StudentEmail, "Expected student email to be in the list of students")
}

func TestGetCommonStudents(t *testing.T) {
	// Set-up Test Data
	createAndLoad()
	registries := []models.Registry{
		{TeacherEmail: "teacherken@gmail.com", StudentEmail: "studentjon@gmail.com"},
		{TeacherEmail: "teacherken@gmail.com", StudentEmail: "studenthon@gmail.com"},
		{TeacherEmail: "teacherken@gmail.com", StudentEmail: "studentunderkenonly@gmail.com"},
	}
	models.DB.Create(&registries)

	request, _ := http.NewRequest("GET", "/api/commonstudents?teacher=teacherken@gmail.com", nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/commonstudents", controllers.GetCommonStudents).Methods("GET")
	router.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")

	assert.JSONEq(t, `{"students": ["studenthon@gmail.com", "studentjon@gmail.com", "studentunderkenonly@gmail.com"]}`, response.Body.String())
}

func TestGetCommonStudentsWith2Teachers(t *testing.T) {
	// Set-up Test Data
	createAndLoad()
	registries := []models.Registry{
		{TeacherEmail: "teacherken@gmail.com", StudentEmail: "studentjon@gmail.com"},
		{TeacherEmail: "teacherken@gmail.com", StudentEmail: "studenthon@gmail.com"},
		{TeacherEmail: "teacherken@gmail.com", StudentEmail: "studentunderkenonly@gmail.com"},
	}
	models.DB.Create(&registries)

	request, _ := http.NewRequest("GET", "/api/commonstudents?teacher=teacherken@gmail.com&teacher=teacherjoe@gmail.com", nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/commonstudents", controllers.GetCommonStudents).Methods("GET")
	router.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")

	assert.JSONEq(t, `{"students": ["studenthon@gmail.com", "studentjon@gmail.com"]}`, response.Body.String())
}

func TestCheckStudentSuspended(t *testing.T) {
	// Set-up Test Data
	createAndLoad()

	// Case when: Student is suspended
	models.DB.Create(&models.Student{Email: "suspendedstudent@gmail.com", Suspended: 1})
	isSuspended := controllers.CheckStudentSuspended("suspendedstudent@gmail.com")
	assert.True(t, isSuspended, "Expect student to be suspended")

	// Case when: Student is not suspended
	models.DB.Create(&models.Student{Email: "notsuspendedstudent@gmail.com", Suspended: 0})
	isSuspended = controllers.CheckStudentSuspended("notsuspendedstudent@gmail.com")
	assert.False(t, isSuspended, "Expect student to not be suspended")

	// Case when: Student does not exist
	isSuspended = controllers.CheckStudentSuspended("nonexistentstudent@gmail.com")
	assert.True(t, isSuspended, "Expect student to be suspended")
}

func TestSuspendStudent(t *testing.T) {
	requestBody := controllers.SuspendStudentRequest{
		Student: "studentjon@gmail.com",
	}

	// Set-up Test Data
	createAndLoad()

	// Assert that initial student is not suspended
	var initStudent models.Student
	models.DB.Where("email = ?", requestBody.Student).First(&initStudent)
	assert.Equal(t, 0, initStudent.Suspended, "Expect suspended field to be equal to 0")

	jsonStr, _ := json.Marshal(requestBody)
	request, _ := http.NewRequest("POST", "/api/suspend", bytes.NewBuffer(jsonStr))
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/suspend", controllers.SuspendStudent).Methods("POST")
	router.ServeHTTP(response, request)
	assert.Equal(t, 204, response.Code, "OK response is expected")

	// Assert that the student is suspended
	var student models.Student
	models.DB.Where("email = ?", requestBody.Student).First(&student)
	assert.Equal(t, 1, student.Suspended, "Expect suspended field to be equal to 1")
}

func TestRetrieveNotification(t *testing.T) {
	requestBody := controllers.GetStudentsWithNotificationRequest{
		Teacher:      "teacherken@gmail.com",
		Notification: "Hello students! @studenttom@gmail.com",
	}

	// Set-up Test Data
	createAndLoad()
	models.DB.Model(&models.Student{}).Where("email = ?", "studentjon@gmail.com").Update("suspended", 1)
	registries := []models.Registry{
		{TeacherEmail: "teacherken@gmail.com", StudentEmail: "studentjon@gmail.com"},
		{TeacherEmail: "teacherken@gmail.com", StudentEmail: "studenthon@gmail.com"},
		{TeacherEmail: "teacherken@gmail.com", StudentEmail: "studentunderkenonly@gmail.com"},
	}
	models.DB.Create(&registries)

	jsonStr, _ := json.Marshal(requestBody)
	request, _ := http.NewRequest("POST", "/api/retrievefornotifications", bytes.NewBuffer(jsonStr))
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/retrievefornotifications", controllers.GetStudentsWithNotification).Methods("POST")
	router.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")

	assert.JSONEq(t, `{"recipients": ["studenttom@gmail.com", "studenthon@gmail.com", "studentunderkenonly@gmail.com"]}`, response.Body.String())
}

func TestRetrieveNotificationWithoutMentions(t *testing.T) {
	requestBody := controllers.GetStudentsWithNotificationRequest{
		Teacher:      "teacherken@gmail.com",
		Notification: "Hello students!",
	}

	// Set-up Test Data
	createAndLoad()
	models.DB.Model(&models.Student{}).Where("email = ?", "studentjon@gmail.com").Update("suspended", 1)
	registries := []models.Registry{
		{TeacherEmail: "teacherken@gmail.com", StudentEmail: "studentjon@gmail.com"},
		{TeacherEmail: "teacherken@gmail.com", StudentEmail: "studenthon@gmail.com"},
		{TeacherEmail: "teacherken@gmail.com", StudentEmail: "studentunderkenonly@gmail.com"},
	}
	models.DB.Create(&registries)

	jsonStr, _ := json.Marshal(requestBody)
	request, _ := http.NewRequest("POST", "/api/retrievefornotifications", bytes.NewBuffer(jsonStr))
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/retrievefornotifications", controllers.GetStudentsWithNotification).Methods("POST")
	router.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")

	assert.JSONEq(t, `{"recipients": ["studenthon@gmail.com", "studentunderkenonly@gmail.com"]}`, response.Body.String())
}
