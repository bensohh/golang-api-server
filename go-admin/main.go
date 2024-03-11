package main

import (
	"fmt"
	"net/http"

	"github.com/bensohh/go-admin/controllers"
	"github.com/bensohh/go-admin/models"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	handler := controllers.New()

	fmt.Println("Connecting to Database...")

	models.ConnectDatabase()

	fmt.Println("Database Connected")

	err := http.ListenAndServe(":3333", handler)
	if err != nil {
		fmt.Println("Server died...")
	}

}
