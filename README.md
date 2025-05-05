# Habit and Productivity Tracker

A microservices-based application for tracking habits and productivity.

## Project Structure

The project consists of two main microservices:

### 1. User Service
- Handles user registration and authentication
- Manages user profiles
- Provides JWT-based authentication for other services
- Runs on port 8080

### 2. Tracker Service
- Manages habit tracking and statistics
- Provides analytics and motivation
- Runs on port 8081

## Setup

### Prerequisites
- Go 1.24.2 or later
- PostgreSQL
- Docker (optional)

### Database Setup
1. Create a PostgreSQL database for each service
2. Update the configuration files with your database credentials

### Running the Services

#### User Service
```bash
cd user-service
go mod tidy
go run cmd/api/main.go
```

#### Tracker Service
```bash
cd tracker-service
go mod tidy
go run cmd/api/main.go
```

## API Endpoints

### User Service
- POST /register - User registration
- POST /login - User login (returns JWT)
- GET /me - Get current user information

### Tracker Service
- POST /habits - Create a new habit
- GET /habits - List all habits
- POST /habits/{id}/track - Mark habit completion
- GET /habits/{id}/stats - Get habit statistics
- GET /habits/{id}/motivation - Get motivational content

## Development

Each service is independently deployable and communicates via HTTP. The services use JWT for authentication between them.

## License

MIT 