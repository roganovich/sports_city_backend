package main

import (
	"flag"
	_ "goland_api/docs"
	"goland_api/pkg/cmd"
	"goland_api/pkg/database"
	"goland_api/pkg/handlers"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/swaggo/http-swagger"
)

// @title My Golang API
// @description This is a sample server.
// @version 1.0
// @host localhost:8080
// @BasePath /api
func main() {
	// Загружаем .env файл
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// InitDB
	dataSourceName := os.Getenv("DATABASE_URL")
	database.InitDB(dataSourceName)

	// Запуск консольных команд
	consoleName := flag.String("consoleName", "Default", "Console Name")

	if *consoleName == "importFields" {
		log.Println("Run Console ", *consoleName)
		cmd.RunImportFields()
	} else if *consoleName != "Default" {
		log.Println("Unknown Console ", *consoleName)
	}

	// Регистрация маршрутов
	router := mux.NewRouter()

	// Применяем middleware CORS ко всем роутам
	router.Use(handlers.CORS)

	// Swagger
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Участники
	router.HandleFunc("/api/users", handlers.AuthAdminMiddleware(handlers.GetUsers())).Methods("GET")
	router.HandleFunc("/api/users/{id}", handlers.AuthAdminMiddleware(handlers.GetUser())).Methods("GET")

	// Кабинет
	router.HandleFunc("/api/auth/info", handlers.AuthMiddleware(handlers.InfoUser())).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/auth/create", handlers.CreateUser()).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/update", handlers.AuthMiddleware(handlers.UpdateUser())).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/auth/login", handlers.Login()).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/refresh", handlers.AuthMiddleware(handlers.Refresh())).Methods("POST", "OPTIONS")
	//router.HandleFunc("/api/auth", handlers.DeleteUser()).Methods("DELETE")

	// Команды
	router.HandleFunc("/api/teams", handlers.GetTeams()).Methods("GET")
	router.HandleFunc("/api/teams/{id}", handlers.GetTeam()).Methods("GET")
	router.HandleFunc("/api/teams", handlers.AuthUserMiddleware(handlers.CreateTeam())).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/teams/{id}", handlers.AuthUserMiddleware(handlers.UpdateTeam())).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/teams/{id}", handlers.AuthUserMiddleware(handlers.DeleteTeam())).Methods("DELETE", "OPTIONS")

	// Площадки
	router.HandleFunc("/api/fields", handlers.GetFields()).Methods("GET")
	router.HandleFunc("/api/fields/{id}", handlers.GetField()).Methods("GET")
	router.HandleFunc("/api/fields", handlers.AuthAdminMiddleware(handlers.CreateField())).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/fields/{id}", handlers.AuthAdminMiddleware(handlers.UpdateField())).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/fields/{id}", handlers.AuthAdminMiddleware(handlers.DeleteField())).Methods("DELETE", "OPTIONS")

	// Аренда
	router.HandleFunc("/api/rentals", handlers.GetRentals()).Methods("GET")
	router.HandleFunc("/api/rentals/{id}", handlers.GetRental()).Methods("GET")
	router.HandleFunc("/api/rentals", handlers.AuthMiddleware(handlers.CreateRental())).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/rentals/{id}", handlers.AuthMiddleware(handlers.DeleteField())).Methods("DELETE", "OPTIONS")

	// Media
	router.HandleFunc("/api/media/preloader", handlers.AuthMiddleware(handlers.Preloader())).Methods("POST", "OPTIONS")

	// Адресса
	router.HandleFunc("/api/address/suggest", handlers.AuthMiddleware(handlers.SuggestAddress())).Methods("POST", "OPTIONS")

	//start server
	log.Fatal(http.ListenAndServe(":8000", handlers.JsonContentTypeMiddleware(router)))
}
