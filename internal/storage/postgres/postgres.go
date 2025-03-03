package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Foreground-Eclipse/transferer/config"
	_ "github.com/lib/pq"
)

// docker run --name exchanger -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=Tatsh -e POSTGRES_DB=exchanger -d postgres
type Storage struct {
	db *sql.DB
}

func New(cfg *config.Config) (*Storage, error) {
	const op = "storage.postgres.New"

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SSLMode)
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	db.SetMaxOpenConns(150)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(time.Minute * 5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) InitUserSchema() error {
	const op = "storage.postgres.InitUserSchema"
	query := `
	CREATE TABLE IF NOT EXISTS users (
    ID INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,         
    password_hash VARCHAR(255) NOT NULL,          
    email VARCHAR(255) UNIQUE
);`

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	return nil
}

func (s *Storage) InitWalletSchema() error {
	const op = "storage.postgres.InitWalletSchema"
	query := `
	CREATE TABLE IF NOT EXISTS wallets (
    ID INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(ID) ON DELETE CASCADE, 
    currency VARCHAR(3) NOT NULL,                  
    balance NUMERIC(19, 8) NOT NULL DEFAULT 0,     
    UNIQUE (user_id, currency)
);`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	return nil
}

func (s *Storage) InitExchangeRatesSchema() error {
	const op = "storage.postgres.InitExchangeRatesSchema"
	query := `
	CREATE TABLE IF NOT EXISTS currency (
    from_currency VARCHAR(3) NOT NULL,
    to_currency VARCHAR(3) NOT NULL,
    rate NUMERIC(19, 8) NOT NULL,
    PRIMARY KEY (from_currency, to_currency)
);`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	return nil
}

func (s *Storage) InitExchangeRates() error {
	const op = "storage.postgres.InitExchangeRates"
	query := `
	INSERT INTO currency (from_currency, to_currency, rate)
VALUES
    ('RUB', 'USD', 0.012 + (RANDOM() * 0.001)),   
    ('RUB', 'EUR', 0.011 + (RANDOM() * 0.001)),
    ('RUB', 'RUB', 1.0)   
;`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	return nil
}

func (s *Storage) RegisterUser(username, passwordHash, email string) error {
	const op = "storage.postgres.RegisterUser"
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: begin transaction: %w", op, err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-throw the panic
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				log.Println("Error commiting transaction, rolling back")
				tx.Rollback()
			}
		}
	}()
	var userID int
	queryInsertUser := `
    INSERT INTO users (username, password_hash, email)
    VALUES ($1, $2, $3)
    RETURNING ID;
`

	err = tx.QueryRow(queryInsertUser, username, passwordHash, email).Scan(&userID)
	if err != nil {
		return fmt.Errorf("%s: insert user: %w", op, err)
	}
	currencies := []string{"USD", "EUR", "RUB"}

	queryInsertWallet := `
        INSERT INTO wallets (user_id, currency)
        VALUES ($1, $2);
    `
	for _, currency := range currencies {
		_, err = tx.Exec(queryInsertWallet, userID, currency)
		if err != nil {
			return fmt.Errorf("%s: insert wallet for %s: %w", op, currency, err)
		}
	}
	return nil
}

func (s *Storage) DoesUserExists(username string) (bool, error) {
	const op = "storage.postgres.DoesUserExists"

	query := `
	SELECT COUNT (*) FROM USERS WHERE username = $1`
	var count int

	err := s.db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return true, fmt.Errorf("%s :%w", op, err)
	}

	if count != 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func (s *Storage) DoesEmailExists(email string) (bool, error) {
	const op = "storage.postgres.DoesEmailExists"

	query := `
	SELECT COUNT (*) FROM USERS WHERE email = $1`
	var count int

	err := s.db.QueryRow(query, email).Scan(&count)
	if err != nil {
		return true, fmt.Errorf("%s :%w", op, err)
	}

	if count != 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func (s *Storage) GetUsersPassHash(username string) (string, error) {
	const op = "storage.postgres.GetUsersPassHash"

	query := `
	SELECT password_hash FROM users WHERE username = $1`
	var passhash string

	err := s.db.QueryRow(query, username).Scan(&passhash)
	if err != nil {
		return "", fmt.Errorf("%s :%w", op, err)
	}

	return passhash, nil
}

func (s *Storage) GetUserBalance(username string) (map[string]float64, error) {
	query := `
		SELECT
			w.currency,
			SUM(w.balance)
		FROM
			wallets w
			JOIN users u ON w.user_id = u.id
		WHERE
			u.username = $1
		GROUP BY
			w.currency;
	`

	rows, err := s.db.Query(query, username)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	balances := make(map[string]float64)
	for rows.Next() {
		var currency string
		var balance float64
		err := rows.Scan(&currency, &balance)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		balances[currency] = balance
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return balances, nil
}

func (s *Storage) UpdateUsersBalance(username, currency string, amount float64) error {
	const op = "storage.postgres.UpdateUsersBalance"

	queryUserCheck := `SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)`
	var userExists bool
	err := s.db.QueryRow(queryUserCheck, username).Scan(&userExists)
	if err != nil {
		return fmt.Errorf("%s: failed to check if user exists: %w", op, err)
	}
	if !userExists {
		return fmt.Errorf("%s: user %s does not exist", op, username)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				log.Println("Error commiting transaction, rolling back")
				tx.Rollback()
			}
		}
	}()

	var userID int
	queryGetUserID := `SELECT id FROM users WHERE username = $1`
	err = tx.QueryRow(queryGetUserID, username).Scan(&userID)
	if err != nil {
		return fmt.Errorf("%s: failed to get user_id: %w", op, err)
	}

	queryUpdate := `
    UPDATE wallets
    SET balance = balance + $1
    WHERE user_id = $2 AND currency = $3
	`
	result, err := tx.Exec(queryUpdate, amount, userID, currency)
	if err != nil {
		return fmt.Errorf("%s: failed to update wallet: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: No wallet found for user %s and currency %s", op, username, currency)
	}

	log.Printf("Updated wallet for user %s in currency %s by %f\n", username, currency, amount)

	return nil
}
