# Tickitz - Cinema Ticket Booking Backend API

## 📖 Project Overview

**Tickitz Backend** is a RESTful API server that powers the Tickitz cinema ticket booking application, handling user authentication, movie management, seat reservations, and payment processing.

**Technologies Used:**
- Go 1.24.6 (Programming language)
- Gin Gonic (Web framework)
- PostgreSQL (Database with pgx driver)
- JWT (golang-jwt/jwt/v5)
- Redis (Caching)
- Swagger (API documentation)

**Key Features:**
- 🔐 **User Authentication** - Registration, login, password reset
- 🎬 **Movie Management** - CRUD operations for movies and showtimes
- 🎫 **Booking System** - Seat reservation and availability management
- 💳 **Payment Integration** - Transaction processing and validation
- 🎟️ **Ticket Generation** - E-ticket creation and management
- 📊 **Admin Dashboard** - Cinema and show management

## 🚀 Instructions

**Environment Requirements:**
- Go 1.24.6+
- PostgreSQL 13+
- Redis

**Installation & Setup:**
```bash
# Clone repository
git clone https://github.com/radifan9/tickitz-ticketing-backend.git
cd tickitz-ticketing-backend

# Download dependencies
go mod download

# Setup environment variables
cp .env.example .env
# Edit .env with your database and API configurations

# Run database migrations (if available)
go run cmd/migrate/main.go

# Start development server
go run main.go

# Build for production
go build -o tickitz-backend
./tickitz-backend
```

**Environment Variables:**
```bash
# Database Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=pg_user_example
POSTGRES_PASSWORD=your_postgres_password_example
POSTGRES_DB=db_name_example

# JWT Configuration
JWT_ISSUER=jwt_issuer_example
JWT_SECRET=your_super_secret_jwt_key_example

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6378
REDIS_USER=rdb_user_example
REDIS_PASSWORD=your_redis_password_example
```

## 📋 API Documentation

### Authentication Endpoints
```http
POST   /api/v1/auth/register    # User registration
POST   /api/v1/auth/login       # User login
DELETE /api/v1/auth/logout      # User logout (requires auth)
```

### User Profile Endpoints
```http
GET    /api/v1/users/profile    # Get user profile (requires auth)
PATCH  /api/v1/users/profile    # Update user profile (requires auth)
PATCH  /api/v1/users/password   # Change password (requires auth)
```

### Movies Endpoints
```http
GET    /api/v1/movies/          # Get filtered movies
GET    /api/v1/movies/:id       # Get movie details
GET    /api/v1/movies/upcoming  # Get upcoming movies
GET    /api/v1/movies/popular   # Get popular movies
```


### Static Files
```http
GET    /api/v1/img/*            # Serve static images
```

**Response Format:**
```json
{
  "success": true,
  "status": 200,
  "data": {},
}
```

**Error Response Format:**
```json
{
  "success": false,
  "status": 400,
  "error": "error message description"
}
```

**Authentication:**
- Protected endpoints require `Authorization: Bearer <token>` header
- Token blacklist implemented for secure logout
- Role-based access control (admin/user roles)

## ℹ️ Other Information

**License:** MIT

**Contact:** 
- GitHub: [@radifan9](https://github.com/radifan9)

**Related Project:**
- [Tickitz Frontend](https://github.com/radifan9/tickitz-ticketing-react) - React client application

**API Base URL:** 
- Development: `http://localhost:8080/api/v1`
- **Swagger Documentation**: `http://localhost:8080/swagger/index.html`

**Project Structure:**
```
├── cmd/                 # cmd folder
│   └── main.go          # Application entry point
├── internal/            # Private application code
│   ├── handlers/        # HTTP handlers
│   ├── repositories/    # Data access layer
│   ├── models/          # Data models
│   └── middleware/      # HTTP middleware
├── pkg/                 # Public libraries
├── docs/                # Swagger documentation
└── migrations/          # Database migrations

```