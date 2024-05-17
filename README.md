# Chirpy

Chirpy is a microblogging application built with Go, offering functionalities similar to Twitter. It utilizes a RESTful API architecture and includes features for managing chirps (posts) and users.

## Features

- CRUD functionalities for chirps
- User management
- Password-based authentication
- Bcrypt hashing for password storage
- JSON Web Tokens (JWT) for authentication

## Technologies Used

- Go (Golang)
- RESTful API
- Bcrypt hashing
- JSON Web Tokens (JWT)

## Installation

1. Clone the repository:
```
git clone github.com/friday1602/chirpy
```
2. Install dependencies:
```
go mod download
```
3. Configure environment variables:

- Edit `.env` file with your configurations.

4. Build and run the application:
```
go build -o chirpy && ./chirpy
```

## Usage

1. Create a new user using `POST /api/users`.
2. Login with your credentials using `/api/login` to obtain a JWT token.
3. Use the obtained JWT token for authentication in subsequent requests to protected endpoints.

