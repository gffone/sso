package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"sso/internal/domain/models"
	"sso/internal/storage"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "pwd"
	dbname   = "test"
)

type Storage struct {
	db *sql.DB
}

func New() (*Storage, error) {
	const op = "storage.postgres.New"

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	fmt.Println("successfully connected to DB")
	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.postgres.SaveUser"

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO users (email, pass_hash) VALUES ($1, $2);")
	txStmt := tx.Stmt(stmt)

	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return 0, fmt.Errorf("%s: %w", op, fmt.Errorf(fmt.Sprintf("unable to rollback %v", rollbackErr)))
		}
		return 0, fmt.Errorf("%s: %w , tx rollback", op, err)
	}

	_, err = txStmt.ExecContext(ctx, email, passHash)
	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return 0, fmt.Errorf("%s: %w", op, fmt.Errorf(fmt.Sprintf("unable to rollback %v", rollbackErr)))
		}

		return 0, fmt.Errorf("%s: %w , tx rollback", op, err)
	}

	row := tx.QueryRowContext(ctx, "SELECT currval(pg_get_serial_sequence('users', 'id'));")

	var uid int64

	err = row.Scan(&uid)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if comErr := tx.Commit(); comErr != nil {
		return 0, fmt.Errorf("%s: %w", op, fmt.Sprintf("unable to commit %v", comErr))
	}

	return uid, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.User"

	var user models.User

	stmt, err := s.db.Prepare("SELECT id, email, pass_hash FROM users WHERE email = $1")
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, email)

	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, storage.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.postgres.IsAdmin"

	stmt, err := s.db.Prepare("SELECT is_admin FROM users WHERE id = $1")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	row := stmt.QueryRowContext(ctx, userID)

	var isAdmin bool

	err = row.Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

func (s *Storage) App(ctx context.Context, appID int64) (models.App, error) {
	const op = "storage.postgres.App"

	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = $1")
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, appID)

	var app models.App

	err = row.Scan(&app.ID, &app.Name, &app.Secret)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}
