package models

// CREATE TABLE IF NOT EXISTS users (
//     ID INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     username VARCHAR(255) UNIQUE NOT NULL,
//     password_hash VARCHAR(255) NOT NULL,
//     email VARCHAR(255) UNIQUE,
//     created_at TIMESTAMPTZ DEFAULT NOW(),
//     updated_at TIMESTAMPTZ DEFAULT NOW()

type User struct {
	ID           int
	Username     string
	PasswordHash string
	Email        string
}
