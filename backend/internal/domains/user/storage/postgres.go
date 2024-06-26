package storage

import (
	"fmt"
	"os"

	"github.com/Corray333/univer_cs/internal/domains/user/types"
	"github.com/Corray333/univer_cs/pkg/server/auth"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type UserStorage struct {
	db    *sqlx.DB
	redis *redis.Client
}

// New creates a new storage and tables
func NewStorage(db *sqlx.DB) *UserStorage {
	return &UserStorage{
		db: db,
		redis: redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_ADDR"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0,
		}),
	}
}

// InsertUser inserts a new user into the database and returns the id
func (s *UserStorage) InsertUser(user types.User, agent string) (int, string, error) {
	passHash, err := auth.Hash(user.Password)
	if err != nil {
		return -1, "", err
	}
	user.Password = passHash

	tx, err := s.db.Begin()
	if err != nil {
		return -1, "", err
	}

	rows := tx.QueryRow(`
		INSERT INTO users (username, email, password, avatar) VALUES ($1, $2, $3, $4) RETURNING user_id;
	`, user.Username, user.Email, user.Password, "http://localhost:3001/images/avatars/default_avatar.png")

	if err := rows.Scan(&user.ID); err != nil {
		return -1, "", err
	}

	refresh, err := auth.CreateToken(user.ID, auth.RefreshTokenLifeTime)
	if err != nil {
		return -1, "", err
	}

	if err := s.SetRefreshToken(user.ID, agent, refresh); err != nil {
		tx.Rollback()
		return -1, "", err
	}

	// _, err = s.db.Queryx(`
	// 	INSERT INTO user_token (user_id, token) VALUES ($1, $2);
	// `, user.ID, refresh)
	// if err != nil {
	// 	return -1, "", err
	// }

	tx.Commit()

	return user.ID, refresh, nil
}

// LoginUser checks if the user exists and the password is correct
func (s *UserStorage) LoginUser(user types.User, agent string) (int, string, error) {
	password := user.Password

	tx, err := s.db.Beginx()
	if err != nil {
		return -1, "", err
	}

	rows := tx.QueryRow(`
		SELECT user_id, password FROM users WHERE email = $1;
	`, user.Email)

	if err := rows.Scan(&user.ID, &user.Password); err != nil {
		return -1, "", err
	}
	if !auth.Verify(user.Password, password) {
		return -1, "", fmt.Errorf("invalid password")
	}

	// Auto update refresh token
	refresh, err := auth.CreateToken(user.ID, auth.RefreshTokenLifeTime)
	if err != nil {
		return -1, "", err
	}

	if err := s.SetRefreshToken(user.ID, agent, refresh); err != nil {
		tx.Rollback()
		return -1, "", err
	}

	// _, err = tx.Queryx(`
	// 	INSERT INTO user_token (user_id, token) VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE SET token = $3;
	// `, user.ID, refresh, refresh)
	// if err != nil {
	// 	return -1, "", err
	// }

	tx.Commit()

	return user.ID, refresh, nil
}

func (s *UserStorage) SelectUser(id string) (types.User, error) {
	var user types.User
	rows, err := s.db.Queryx(`
		SELECT * FROM users WHERE user_id = $1;
	`, id)
	if err != nil {
		return user, err
	}
	if !rows.Next() {
		return user, fmt.Errorf("user not found")
	}
	if err := rows.StructScan(&user); err != nil {
		return user, err
	}
	user.Password = ""
	return user, nil
}

func (s *UserStorage) UpdateUser(user types.User) error {
	fmt.Println()
	fmt.Println(user)
	fmt.Println()
	_, err := s.db.Queryx(`
		UPDATE users SET username = $1, email = $2, avatar = $3 WHERE user_id = $4;
	`, user.Username, user.Email, user.Avatar, user.ID)
	return err
}
