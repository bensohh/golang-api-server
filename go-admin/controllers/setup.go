package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
)

func New() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/", TestServer).Methods("GET")
	router.HandleFunc("/api/register", RegisterStudents).Methods("POST")
	router.HandleFunc("/api/commonstudents", GetCommonStudents).Methods("GET")
	router.HandleFunc("/api/suspend", SuspendStudent).Methods("POST")
	router.HandleFunc("/api/retrievefornotifications", GetStudentsWithNotification).Methods("POST")

	return router
}
