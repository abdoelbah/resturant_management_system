package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"resturant/controllers"
	"resturant/utils"

	"github.com/go-michi/michi"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/handlers"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
	}

	// Get the env from the environment variable
	connStr := os.Getenv("DATABASE_CONNECTION_STR")
	if connStr == "" {
		log.Fatal("DATABASE_CONNECTION_STR not set in .env file")
	}
	migRoot := os.Getenv("MIGRATIONS_ROOT")
	if migRoot == "" {
		log.Fatal("MIGRATIONS_ROOT not set in .env file")
	}
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		log.Fatal("DOMAIN not set in .env file")
	}

	// Connect to the database
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
	}
	defer db.Close()

	// Set global db variable in controllers
	controllers.SetDB(db)

	// Handle migrations
	mig, err := migrate.New(
		"file://"+GetRootPath("database/migrations"),
		connStr,
	)
	if err != nil {
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
	}
	if err := mig.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		}
		log.Printf("migrations: %s", err.Error())
	}

	// Initialize the router and define routes
	r := michi.NewRouter()
	r.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	r.Route("/customer", func(sub *michi.Router) {
		sub.HandleFunc("POST signup", controllers.Signup)
		sub.HandleFunc("POST login", controllers.Login)
		sub.HandleFunc("PUT update/{id}", controllers.UpdateUser)
		sub.HandleFunc("DELETE delete/{id}", controllers.DeleteUser)
		sub.HandleFunc("GET users", controllers.GetAllUsers)

	})

	r.Route("/admin", func(sub *michi.Router) {
		sub.HandleFunc("POST signup", controllers.AdminSignup)
		sub.HandleFunc("POST login", controllers.AdminLogin)
		sub.HandleFunc("POST add-vendor", controllers.AddVendor)
		sub.HandleFunc("PUT update-vendor/{id}", controllers.UpdateVendor)
		sub.HandleFunc("DELETE delete/{id}", controllers.DeleteVendor)
		sub.HandleFunc("GET list-vendors", controllers.GetAllVendors)
		sub.HandleFunc("GET vendor/{id}", controllers.GetVendorById)

	})

	// Enable CORS
	corsOptions := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // Allow all origins (adjust as needed)
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	fmt.Println("Server running on port 8000 🚀")
	if err := http.ListenAndServe(":8000", corsOptions(r)); err != nil {
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
	}
}

func GetRootPath(dir string) string {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
	}
	absPath := path.Join(path.Dir(ex), dir)
	fmt.Println("Resolved migration path:", absPath) // Debugging line
	return absPath
}
