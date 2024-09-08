package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"resturant/models"
	"resturant/utils"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	_ "github.com/go-michi/michi"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	db          *sqlx.DB
	QB          = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	userColumns = []string{"id", "name", "email", "phone", "password", "created_at", "updated_at"}
)

const customerRoleID = 3

func SetDB(database *sqlx.DB) {
	db = database
}

func Signup(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // Limit to 10 MB
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Invalid form data")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	password := r.FormValue("password")

	if username == "" || password == "" || email == "" || phone == "" {
		utils.HandleError(w, http.StatusBadRequest, "Make sure you fill all fields")
		return
	}

	// Check if the user is already signed up
	query, args, err := QB.Select("id", "email").From("users").Where(squirrel.Eq{"email": email}).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to select user")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	isUserSigned := models.User{}
	if err := db.Get(&isUserSigned, query, args...); err == nil {
		utils.SendJSONResponse(w, http.StatusConflict, "User is already signed up")
		return
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to hash password")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	// Handle file upload for the image
	var imgPath string
	file, handler, err := r.FormFile("img")
	if err == nil {
		defer file.Close()

		imgPath, err = utils.SaveImageFile(file, "users", handler.Filename)
		if err != nil {
			utils.HandleError(w, http.StatusInternalServerError, "Failed to save image")
			log.Fatal(utils.ErrorWithTrace(err, err.Error()))
			return
		}
	} else {
		imgPath = ""
	}

	// Convert backslashes to forward slashes for URI compatibility
	imgPath = strings.ReplaceAll(imgPath, "\\", "/")
	imgURI := fmt.Sprintf("http://localhost:8000/%s", imgPath)

	// Create the new user object
	user := models.User{
		ID:        uuid.New(),
		Name:      username,
		Email:     email,
		Phone:     phone,
		Password:  hashedPassword,
		Img:       imgURI,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert the new user into the database
	query, args, err = QB.Insert("users").
		Columns("id", "name", "phone", "email", "password", "img", "created_at", "updated_at").
		Values(user.ID, user.Name, user.Phone, user.Email, user.Password, user.Img, user.CreatedAt, user.UpdatedAt).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(userColumns, ", "))).
		ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to insert user")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	// Execute the query and scan the result into the user struct
	if err := db.QueryRowx(query, args...).StructScan(&user); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Error creating user: "+err.Error())
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	// Assign the customer role to the new user
	query, args, err = QB.Insert("user_roles").
		Columns("user_id", "role_id").
		Values(user.ID, customerRoleID).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to assign role to user")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	// Execute the role assignment
	if _, err := db.Exec(query, args...); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Error assigning role: "+err.Error())
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	// Send a JSON response with the created user information
	utils.SendJSONResponse(w, http.StatusCreated, user)
}

func Login(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Invalid form data")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	// Get email and password from form fields
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Check if both fields are filled
	if email == "" || password == "" {
		utils.HandleError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Query to check if the user exists
	var user models.User
	query, args, err := QB.Select("id", "name", "email", "password", "img").From("users").Where(squirrel.Eq{"email": email}).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create query")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	// Execute query to fetch the user from the database
	if err := db.Get(&user, query, args...); err != nil {
		utils.HandleError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Compare the provided password with the hashed password in the database
	if err := utils.CheckPassword(user.Password, password); err != nil {
		utils.HandleError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Successful login: return the user's details (excluding password)
	responseUser := map[string]interface{}{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"img":   user.Img,
	}

	utils.SendJSONResponse(w, http.StatusOK, responseUser)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the URL path parameters
	userID := r.PathValue("id") // Adjust this depending on your router, e.g., Gorilla Mux uses mux.Vars(r)
	if userID == "" {
		utils.HandleError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Parse form data
	err := r.ParseMultipartForm(10 << 20) // Limit to 10 MB for file upload
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Invalid form data")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	// Fetch the current user from the database
	var user models.User
	query, args, err := QB.Select("id", "name", "img").From("users").Where(squirrel.Eq{"id": userID}).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create query")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	if err := db.Get(&user, query, args...); err != nil {
		utils.HandleError(w, http.StatusNotFound, "User not found")
		return
	}

	// Get the new username and image
	newUsername := r.FormValue("username")
	file, handler, err := r.FormFile("img") // New image file

	var newImgPath string

	// Handle image replacement if a new image is provided
	if err == nil {
		defer file.Close()

		// Delete the old image from the uploads folder
		if user.Img != "" {
			oldImagePath := strings.ReplaceAll(user.Img, "http://localhost:8000/", "") // Strip the base URL
			if err := os.Remove(oldImagePath); err != nil {
				utils.HandleError(w, http.StatusInternalServerError, "Failed to delete old image")
				return
			}
		}

		// Save the new image
		newImgPath, err = utils.SaveImageFile(file, "users", handler.Filename)
		if err != nil {
			utils.HandleError(w, http.StatusInternalServerError, "Failed to save new image")
			log.Fatal(utils.ErrorWithTrace(err, err.Error()))
			return
		}

		// Construct the new image URI
		newImgPath = fmt.Sprintf("http://localhost:8000/%s", strings.ReplaceAll(newImgPath, "\\", "/"))
	} else {
		// If no new image is uploaded, keep the old image path
		newImgPath = user.Img
	}

	// Update the user data in the database
	updateQuery, args, err := QB.Update("users").
		Set("name", newUsername).
		Set("img", newImgPath).
		Where(squirrel.Eq{"id": userID}).
		ToSql()

	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create update query")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	// Execute the update query
	if _, err := db.Exec(updateQuery, args...); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to update user")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	// Return the updated user details
	updatedUser := map[string]interface{}{
		"id":   user.ID,
		"name": newUsername,
		"img":  newImgPath,
	}

	utils.SendJSONResponse(w, http.StatusOK, updatedUser)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the URL path or query parameters
	userID := r.PathValue("id") // Or use path param based on your router
	if userID == "" {
		utils.HandleError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Fetch the user's details (including image path) before deletion
	var user models.User
	query, args, err := QB.Select("id", "img").From("users").Where(squirrel.Eq{"id": userID}).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create query")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	// Execute query to fetch the user from the database
	if err := db.Get(&user, query, args...); err != nil {
		utils.HandleError(w, http.StatusNotFound, "User not found")
		return
	}

	// Delete the user's image from the uploads folder
	if user.Img != "" {
		imgPath := strings.ReplaceAll(user.Img, "http://localhost:8000/", "") // Strip the base URL
		if err := os.Remove(imgPath); err != nil {
			utils.HandleError(w, http.StatusInternalServerError, "Failed to delete user image")
			log.Fatal(utils.ErrorWithTrace(err, err.Error()))
			return
		}
	}

	// Delete the user from the user_roles table
	query, args, err = QB.Delete("user_roles").Where(squirrel.Eq{"user_id": userID}).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create role deletion query")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	if _, err := db.Exec(query, args...); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to delete user roles")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	// Delete the user from the users table
	query, args, err = QB.Delete("users").Where(squirrel.Eq{"id": userID}).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create user deletion query")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	if _, err := db.Exec(query, args...); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to delete user")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	// Return a successful response
	utils.SendJSONResponse(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User

	query, args, err := QB.Select("id", "name", "email", "phone", "img", "created_at", "updated_at").From("users").ToSql()

	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create query")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	if err := db.Select(&users, query, args...); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to fetch users")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, users)
}
