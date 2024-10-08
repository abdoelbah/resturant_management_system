

# Sadeen Internship - Restaurant Management System - Backend

This project is the backend for the **Restaurant Management System**, developed during the **Sadeen Summer Internship 2024**. It provides APIs for managing vendors, customers, and admins. The backend is built with **Go**, **PostgreSQL**, and **Michi** routing, using a layered architecture.

## Project Overview

The backend consists of RESTful APIs to handle various operations, including:

- **Admin Operations**: Admin signup, login, vendor management (add, update, delete).
- **Customer Operations**: Customer signup, login, profile update, and deletion.
- **Vendor Management**: List all vendors, get vendor details, and handle vendor-specific data.
- **Role-based Access Control (RBAC)**: Assign roles to users (admin, vendor, customer).

## Getting Started

### Prerequisites

- **Go** version 1.23+
- **PostgreSQL** for the database
- **Go modules** for dependency management

### Setting Up the Project

1. **Clone the repository**:
   ```bash
   git clone https://github.com/abdoelbah/resturant_management_system_backend.git
   cd sadeen-restaurant-backend
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Set up environment variables**:
   Create a `.env` file and add the following:
   ```bash
   DATABASE_CONNECTION_STR=postgres://youruser:yourpassword@localhost:5432/yourdatabase?sslmode=disable
   MIGRATIONS_ROOT=database/migrations
   DOMAIN=http://localhost:8000
   ```

4. **Run migrations**:
   Migrations are used to set up the database schema. Ensure the migrations are executed with the following command:
   ```bash
   go run main.go
   ```

5. **Run the server**:
   Start the backend server:
   ```bash
   go run main.go
   ```

6. **Access the server**:
   The server will run on `http://localhost:8000`. API routes are prefixed with `/customer` for customer operations and `/admin` for admin operations.

## Project Structure

```
src/
├── controllers/        # Contains admin, customer, and vendor controllers
├── database/           # Database migrations
├── models/             # Structs defining the database schema (User, Role, etc.)
├── uploads/            # Stores uploaded images
├── utils/              # Utility functions (e.g., password hashing, image handling)
└── main.go             # Main entry point of the application
```

## API Endpoints

### Admin Endpoints
- `POST /admin/signup`: Admin signup with image upload.
- `POST /admin/login`: Admin login.
- `POST /admin/add-vendor`: Add a new vendor.
- `PUT /admin/update-vendor/{id}`: Update an existing vendor.
- `DELETE /admin/delete/{id}`: Delete a vendor.
- `GET /admin/list-vendors`: List all vendors.
- `GET /admin/vendor/{id}`: Get vendor details by ID.

### Customer Endpoints
- `POST /customer/signup`: Customer signup with image upload.
- `POST /customer/login`: Customer login.
- `PUT /customer/update/{id}`: Update customer details.
- `DELETE /customer/delete/{id}`: Delete a customer.
- `GET /customer/users`: List all customers.

## Database

- The project uses **PostgreSQL** as the database.
- Migrations are managed using **golang-migrate** to ensure the database schema is versioned and easy to update.

## Notes

This project is part of the **Sadeen Summer Internship 2024** and was designed to provide a complete system for managing restaurants with admins, vendors, and customers.
