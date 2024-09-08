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
	"github.com/google/uuid"
)

const adminRoleID = 1
const vendorRoleID = 2

func AdminSignup(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // Limit to 10 MB for file upload
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Invalid form data")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	// Extract form data
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	phone := r.FormValue("phone")

	// Validate required fields
	if username == "" || email == "" || password == "" || phone == "" {
		fmt.Printf("username: %v, email: %v, password: %v phone: %v", username, email, password, phone)
		utils.HandleError(w, http.StatusBadRequest, "make sure u fill all fields")
		return
	}

	// Check if the admin already exists
	query, args, err := QB.Select("id", "email").From("users").Where(squirrel.Eq{"email": email}).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "enternal server error")
		log.Fatal(utils.ErrorWithTrace(err, err.Error()))
		return
	}

	existingUser := models.User{}
	if err := db.Get(&existingUser, query, args...); err == nil {
		utils.SendJSONResponse(w, http.StatusConflict, "Admin with this email already exists")
		return
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to hash password")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Handle file upload for the admin's image
	var imgPath string
	file, handler, err := r.FormFile("img")
	if err == nil {
		defer file.Close()

		// Save the new image in the uploads/admins directory
		imgPath, err = utils.SaveImageFile(file, "admins", handler.Filename)
		if err != nil {
			utils.HandleError(w, http.StatusInternalServerError, "Failed to save image")
			utils.ErrorWithTrace(err, err.Error())
			return
		}
		// Convert backslashes to forward slashes for URI compatibility
		imgPath = strings.ReplaceAll(imgPath, "\\", "/")
	} else {
		imgPath = ""
	}

	// Construct the image URI
	imgURI := fmt.Sprintf("http://localhost:8000/%s", imgPath)

	// Create a new admin user
	user := models.User{
		ID:        uuid.New(),
		Name:      username,
		Email:     email,
		Phone:     phone, // Include the phone field
		Password:  hashedPassword,
		Img:       imgURI, // Store the image URI in the database
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert the new admin into the users table
	query, args, err = QB.Insert("users").
		Columns("id", "name", "email", "phone", "password", "img", "created_at", "updated_at").
		Values(user.ID, user.Name, user.Email, user.Phone, user.Password, user.Img, user.CreatedAt, user.UpdatedAt).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(userColumns, ", "))).
		ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create admin")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Execute the query and scan the result into the user struct
	if err := db.QueryRowx(query, args...).StructScan(&user); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Error creating admin: "+err.Error())
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Assign the 'admin' role to the user in the user_roles table
	query, args, err = QB.Insert("user_roles").
		Columns("user_id", "role_id").
		Values(user.ID, adminRoleID).
		ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to assign admin role")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	if _, err := db.Exec(query, args...); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Error assigning admin role: "+err.Error())
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Return the newly created admin details
	utils.SendJSONResponse(w, http.StatusCreated, user)
}

func AdminLogin(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Invalid form data")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Get email and password from form fields
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Validate input fields
	if email == "" || password == "" {
		utils.HandleError(w, http.StatusBadRequest, "make sure you fill all fields")
		return
	}

	// Query to check if the user exists
	var user models.User
	query, args, err := QB.Select("id", "name", "email", "password", "img", "phone", "updated_at").From("users").Where(squirrel.Eq{"email": email}).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Internal server error")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Execute query to fetch the user from the database
	if err := db.Get(&user, query, args...); err != nil {
		utils.HandleError(w, http.StatusUnauthorized, "This user is not authorized")
		return
	}

	// Compare the provided password with the hashed password in the database
	if err := utils.CheckPassword(user.Password, password); err != nil {
		utils.HandleError(w, http.StatusUnauthorized, "Passowrd is not correct")
		return
	}

	// Check if the user has the admin role
	var userRole models.UserRole
	query, args, err = QB.Select("role_id").From("user_roles").Where(squirrel.Eq{"user_id": user.ID, "role_id": adminRoleID}).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to check user role")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	if err := db.Get(&userRole, query, args...); err != nil || userRole.RoleID != adminRoleID {
		utils.HandleError(w, http.StatusUnauthorized, "You do not have admin privileges")
		return
	}

	// Successful login: return the admin's details (excluding password)
	responseAdmin := map[string]interface{}{
		"id":        user.ID,
		"name":      user.Name,
		"email":     user.Email,
		"phone":     user.Phone,
		"updatedAt": user.UpdatedAt,
		"image":     user.Img,
	}

	utils.SendJSONResponse(w, http.StatusOK, responseAdmin)
}

func AddVendor(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // Limit to 10 MB for file upload
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Invalid form data")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Extract form data
	username := r.FormValue("username")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	description := r.FormValue("description")

	// Validate required fields
	if username == "" || email == "" || phone == "" || description == "" {
		utils.HandleError(w, http.StatusBadRequest, "Username, email, phone, and description are required")
		return
	}

	// Check if the user already exists
	query, args, err := QB.Select("id", "email").From("users").Where(squirrel.Eq{"email": email}).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Internal server error")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	existingUser := models.User{}
	if err := db.Get(&existingUser, query, args...); err == nil {
		utils.SendJSONResponse(w, http.StatusConflict, "vendor with this email already exists")
		return
	}

	// Handle file upload for the vendor's image
	var imgPath string
	file, fileHeader, err := r.FormFile("img")
	if err == nil {
		defer file.Close()

		// Save the new image in the uploads/vendors directory
		imgPath, err = utils.SaveImageFile(file, "vendors", fileHeader.Filename)
		if err != nil {
			utils.HandleError(w, http.StatusInternalServerError, "Failed to save image")
			log.Fatal(utils.ErrorWithTrace(err, err.Error()))
			return
		}
		// Convert backslashes to forward slashes for URI compatibility
		imgPath = strings.ReplaceAll(imgPath, "\\", "/")
	} else {
		imgPath = ""
	}

	// Construct the image URI
	imgURI := fmt.Sprintf("http://localhost:8000/%s", imgPath)

	// Create a new vendor user in the users table
	user := models.User{
		ID:        uuid.New(),
		Name:      username,
		Email:     email,
		Phone:     phone,
		Img:       imgURI, // Store the image URI in the database
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert the new user into the users table
	query, args, err = QB.Insert("users").
		Columns("id", "name", "email", "phone", "password", "img", "created_at", "updated_at").
		Values(user.ID, user.Name, user.Email, user.Phone, user.Password, user.Img, user.CreatedAt, user.UpdatedAt).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(userColumns, ", "))).
		ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create user")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Execute the query and scan the result into the user struct
	if err := db.QueryRowx(query, args...).StructScan(&user); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Error creating vendor: "+err.Error())
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Assign the 'vendor' role to the user in the user_roles table
	// Assuming 2 is the role ID for 'vendor'
	query, args, err = QB.Insert("user_roles").
		Columns("user_id", "role_id").
		Values(user.ID, vendorRoleID).
		ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to assign vendor role")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	if _, err := db.Exec(query, args...); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Error assigning vendor role: "+err.Error())
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Insert vendor-specific data (description) into the vendors table
	query, args, err = QB.Insert("vendors").
		Columns("vendor_id", "description", "updated_at").
		Values(user.ID, description, time.Now()).
		ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to insert vendor data")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	if _, err := db.Exec(query, args...); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Error inserting vendor data: "+err.Error())
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Return the newly created vendor's details
	utils.SendJSONResponse(w, http.StatusCreated, user)
}

func UpdateVendor(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form data (for file uploads)
	err := r.ParseMultipartForm(10 << 20) // Limit to 10 MB for file upload
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Invalid form data")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Get vendor ID from URL params
	vendorID := r.PathValue("id") // Assuming you are using a router that supports r.PathValue
	if vendorID == "" {
		utils.HandleError(w, http.StatusBadRequest, "Vendor ID is required")
		return
	}

	// Fetch the current vendor data
	var vendor models.User
	query, args, err := QB.Select("id", "name", "img").From("users").Where(squirrel.Eq{"id": vendorID}).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create query")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	if err := db.Get(&vendor, query, args...); err != nil {
		utils.HandleError(w, http.StatusNotFound, "Vendor not found")
		return
	}

	// Extract the new name and description from the form data
	newName := r.FormValue("name")
	newDescription := r.FormValue("description")
	newPhone := r.FormValue("phone")

	// Handle image upload (optional)
	var newImgPath string
	file, fileHeader, err := r.FormFile("img")
	if err == nil {
		defer file.Close()

		// Delete the old image from the uploads directory (if present)
		if vendor.Img != "" {
			oldImagePath := strings.ReplaceAll(vendor.Img, "http://localhost:8000/", "") // Strip base URL
			if err := os.Remove(oldImagePath); err != nil {
				utils.HandleError(w, http.StatusInternalServerError, "Failed to delete old image")
				return
			}
		}

		// Save the new image in the uploads/vendors directory
		newImgPath, err = utils.SaveImageFile(file, "vendors", fileHeader.Filename)
		if err != nil {
			utils.HandleError(w, http.StatusInternalServerError, "Failed to save new image")
			utils.ErrorWithTrace(err, err.Error())
			return
		}
		// Convert backslashes to forward slashes for URI compatibility
		newImgPath = strings.ReplaceAll(newImgPath, "\\", "/")
	} else {
		// If no new image is uploaded, keep the old image path
		newImgPath = vendor.Img
	}

	// Construct the new image URI
	newImgURI := fmt.Sprintf("http://localhost:8000/%s", newImgPath)

	// Update the user data (name, img) in the users table
	updateQuery, args, err := QB.Update("users").
		Set("name", newName).
		Set("img", newImgURI).
		Set("phone", newPhone).
		Where(squirrel.Eq{"id": vendorID}).
		ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create update query")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Execute the update query
	if _, err := db.Exec(updateQuery, args...); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to update vendor data in users table")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Update the vendor-specific data (description) in the vendors table
	updateQuery, args, err = QB.Update("vendors").
		Set("description", newDescription).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"vendor_id": vendorID}).
		ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create update query for vendors table")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Execute the update query for the vendors table
	if _, err := db.Exec(updateQuery, args...); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to update vendor description")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Return the updated vendor details
	updatedVendor := map[string]interface{}{
		"id":          vendor.ID,
		"name":        newName,
		"img":         newImgURI,
		"description": newDescription,
	}

	utils.SendJSONResponse(w, http.StatusOK, updatedVendor)
}

func DeleteVendor(w http.ResponseWriter, r *http.Request) {
	// Get vendor ID from URL params
	vendorID := r.PathValue("id") // Assuming you're using a router that supports r.PathValue
	if vendorID == "" {
		utils.HandleError(w, http.StatusBadRequest, "Vendor ID is required")
		return
	}

	// Fetch the current vendor data to delete the associated image
	var vendor models.User
	query, args, err := QB.Select("id", "img").From("users").Where(squirrel.Eq{"id": vendorID}).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create query to fetch vendor")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	if err := db.Get(&vendor, query, args...); err != nil {
		utils.HandleError(w, http.StatusNotFound, "Vendor not found")
		return
	}

	// Delete the vendor's image from the uploads directory (if it exists)
	if vendor.Img != "" {
		imagePath := strings.ReplaceAll(vendor.Img, "http://localhost:8000/", "") // Strip base URL
		if err := os.Remove(imagePath); err != nil {
			utils.HandleError(w, http.StatusInternalServerError, "Failed to delete vendor image")
			utils.ErrorWithTrace(err, err.Error())
			return
		}
	}

	// Delete vendor data from the vendors table
	query, args, err = QB.Delete("vendors").Where(squirrel.Eq{"vendor_id": vendorID}).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create delete query for vendors")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	if _, err := db.Exec(query, args...); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to delete vendor data from vendors table")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Delete vendor role from the user_roles table
	query, args, err = QB.Delete("user_roles").Where(squirrel.Eq{"user_id": vendorID, "role_id": vendorRoleID}).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create delete query for user roles")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	if _, err := db.Exec(query, args...); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to delete vendor role from user_roles table")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Delete the vendor from the users table
	query, args, err = QB.Delete("users").Where(squirrel.Eq{"id": vendorID}).ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to create delete query for users")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	if _, err := db.Exec(query, args...); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to delete vendor from users table")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Return a success response
	utils.SendJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Vendor and all associated data deleted successfully",
	})
}

func GetAllVendors(w http.ResponseWriter, r *http.Request) {
	// Query to fetch all vendors by joining the users and vendors table
	query, args, err := QB.Select("users.id", "users.name", "users.email", "users.phone", "users.img", "users.created_at", "vendors.description").
		From("users").
		Join("vendors ON users.id = vendors.vendor_id").
		ToSql()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Internal Server Error")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Slice to hold the list of vendors (using models.Vendor)
	var vendors []models.Vendor

	// Execute the query and scan results into the vendors slice
	if err := db.Select(&vendors, query, args...); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to fetch vendors")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Return the list of vendors
	utils.SendJSONResponse(w, http.StatusOK, vendors)
}

func GetVendorById(w http.ResponseWriter, r *http.Request) {
	// Extract the vendor ID from the URL parameters
	vendorId := r.PathValue("id")

	// Query to fetch the vendor by joining the users and vendors table
	query, args, err := QB.Select(
		"users.id",
		"users.name",
		"users.email",
		"users.phone",
		"users.img",
		"users.created_at",
		"vendors.description").
		From("users").
		Join("vendors ON users.id = vendors.vendor_id").
		Where(squirrel.Eq{"users.id": vendorId}).
		ToSql()

	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Internal Server Error")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Struct to hold the vendor data
	var vendor models.Vendor

	// Execute the query and scan results into the vendor struct
	if err := db.Get(&vendor, query, args...); err != nil {
		utils.HandleError(w, http.StatusNotFound, "Vendor not found")
		utils.ErrorWithTrace(err, err.Error())
		return
	}

	// Return the vendor data as JSON
	utils.SendJSONResponse(w, http.StatusOK, vendor)
}
