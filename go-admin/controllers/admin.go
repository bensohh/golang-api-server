package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/bensohh/go-admin/models"
	"github.com/bensohh/go-admin/utils"
)

type CommonStudentsResponse struct {
	Students []string `json:"students"`
}

type RegisterStudentsRequest struct {
	Teacher  string   `json:"teacher"`
	Students []string `json:"students"`
}

type SuspendStudentRequest struct {
	Student string `json:"student"`
}

type GetStudentsWithNotificationRequest struct {
	Teacher      string `json:"teacher"`
	Notification string `json:"notification"`
}

type GetStudentsWithNotificationResponse struct {
	Recipients []string `json:"recipients"`
}

// Checks if the server is running.
func TestServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Server is running")
}

// Checks if a student exists in the db based on their email
func CheckStudentExists(email string) bool {
	var student models.Student
	err := models.DB.Where("email = ?", email).First(&student)
	if err.Error != nil {
		log.Println(err.Error)
		return false
	}
	return true
}

// Registers one/more students to a specified teacher
func RegisterStudents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Retrieve the body parameters
	var bodyParams RegisterStudentsRequest
	err := json.NewDecoder(r.Body).Decode(&bodyParams)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	// Checks if teacher is in db
	var teacher models.Teacher
	result := models.DB.Where("email = ?", bodyParams.Teacher).First(&teacher)

	if result.Error != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid teacher's email")
		return
	}

	// Filter out invalid students (or students not in the db)
	m := make(map[string]bool) // Prevent duplicates
	for _, student := range bodyParams.Students {
		if CheckStudentExists(student) {
			if m[student] {
				continue
			}
			m[student] = true
			newPair := &models.Registry{
				TeacherEmail: teacher.Email,
				StudentEmail: student,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			// Search for teacher_student pair, if it does not exist, we create a new entry in the db
			res := models.DB.Where("teacher_email = ? AND student_email = ?", teacher.Email, student).FirstOrCreate(&newPair)
			if res.Error != nil {
				log.Println(res.Error)
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// Gets the common students given a list of teachers as query params
func GetCommonStudents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Retrieves the query params corresponding to teacher into an array of strings
	teachers := []string(r.URL.Query()["teacher"])

	// Filter teachers query params to ensure that only unique fields exist
	var m = make(map[string]bool)
	var filteredTeachers = []string{}

	for _, teacher := range teachers {
		if m[teacher] {
			continue
		}
		m[teacher] = true
		filteredTeachers = append(filteredTeachers, teacher)
	}

	// Stores the common students' email into an array of strings within the response struct
	var commonStudents CommonStudentsResponse

	// Execute a query on the DB to retrieve all common students
	res := models.DB.Model(&models.Registry{}).Select("student_email").
		Where("teacher_email IN ?", filteredTeachers).
		Group("student_email").
		Having("COUNT(DISTINCT teacher_email) = ?", len(filteredTeachers)).
		Pluck("student_email", &commonStudents.Students)

	if res.Error != nil {
		if len(commonStudents.Students) == 0 {
			json.NewEncoder(w).Encode(commonStudents)
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Error retrieving common students")
		return
	}

	json.NewEncoder(w).Encode(commonStudents)
}

// Updates the student's suspend status to either 0 or 1
func UpdateStudentSuspendStatus(value int, student *models.Student) bool {
	// 0 => Not Suspended, 1 => Suspended
	result := models.DB.Model(&student).Update("suspended", value)
	if result.Error != nil {
		log.Println(result.Error)
		return false
	}
	return true
}

// Suspends a student
func SuspendStudent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var bodyParams SuspendStudentRequest
	err := json.NewDecoder(r.Body).Decode(&bodyParams)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	// Checks if student is in db
	var student models.Student
	res := models.DB.Where("email = ?", bodyParams.Student).First(&student)

	if res.Error != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid student's email")
		return
	}

	// Update the suspended field in student table to 1 => means suspended
	if !UpdateStudentSuspendStatus(1, &student) {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error updating db")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Un-suspend a student
func UnSuspendStudent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var bodyParams SuspendStudentRequest
	err := json.NewDecoder(r.Body).Decode(&bodyParams)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	// Checks if student is in db
	var student models.Student
	res := models.DB.Where("email = ?", bodyParams.Student).First(&student)

	if res.Error != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid student's email")
		return
	}

	// Update the suspended field in student table to 1 => means suspended
	if !UpdateStudentSuspendStatus(0, &student) {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error updating db")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func CheckStudentSuspended(email string) bool {
	var student models.Student
	err := models.DB.Where("email = ?", email).First(&student).Error
	if err != nil {
		return true
	}
	return student.Suspended == 1
}

// Retrieve list of students who can receive a given notification
// remove suspended student
// registered under teacher OR @mentioned
func GetStudentsWithNotification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var bodyParams GetStudentsWithNotificationRequest
	err := json.NewDecoder(r.Body).Decode(&bodyParams)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	// Checks if teacher is in db
	var teacher models.Teacher
	result := models.DB.Where("email = ?", bodyParams.Teacher).First(&teacher)

	if result.Error != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid teacher's email")
		return
	}

	// Retrieve the matching regex for student emails in notification (@ mentioned students)
	regexPattern := regexp.MustCompile(`@\w+@\w+\.\w+`)
	studentEmails := regexPattern.FindAllString(bodyParams.Notification, -1)

	// Retrieve students registered under the teacher
	var registeredStudents []models.Registry
	res := models.DB.Where("teacher_email = ?", teacher.Email).Find(&registeredStudents)

	if res.Error != nil {
		log.Println(res.Error)
		utils.RespondWithError(w, http.StatusInternalServerError, "Error retrieving registered students")
		return
	}

	// Check for duplicates and remove the first '@' character in front of emails
	m := make(map[string]bool)
	var filteredStudents GetStudentsWithNotificationResponse

	// Loop through @ mentioned students and add in those not suspended
	for _, s := range studentEmails {
		if !CheckStudentSuspended(s[1:]) {
			if m[s[1:]] {
				continue
			}
			m[s[1:]] = true
			filteredStudents.Recipients = append(filteredStudents.Recipients, s[1:])
		}
	}

	// Loop through registered students and add in those not suspended
	for _, student := range registeredStudents {
		if !CheckStudentSuspended(student.StudentEmail) {
			if m[student.StudentEmail] {
				continue
			}
			m[student.StudentEmail] = true
			filteredStudents.Recipients = append(filteredStudents.Recipients, student.StudentEmail)
		}
	}

	json.NewEncoder(w).Encode(filteredStudents)
}
