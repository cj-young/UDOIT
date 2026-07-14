package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"rewritetest/internal/shared/apperr"
	"rewritetest/internal/users/internal/domain"
)

type MySQLUserRepository struct {
	db *sql.DB
}

func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{
		db: db,
	}
}

func (r *MySQLUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT id, username, name, preferences
		FROM users
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var (
		userID    int64
		username  string
		name      string
		prefsJSON []byte
	)

	err := row.Scan(&userID, &username, &name, &prefsJSON)
	if err != nil {
		return nil, err
	}

	prefs := domain.Preferences{}
	if len(prefsJSON) > 0 {
		if err := json.Unmarshal(prefsJSON, &prefs); err != nil {
			return nil, err
		}
	}

	user := domain.RehydrateUser(userID, username, name, prefs)

	return &user, nil
}

func (r *MySQLUserRepository) Create(ctx context.Context, user *domain.User) error {
	prefsJSON, err := json.Marshal(user.Preferences())
	if err != nil {
		return err
	}

	query := `
		INSERT INTO users (username, name, preferences)
		VALUES (?, ?, ?)
	`

	var id int64

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Username(),
		user.Name(),
		prefsJSON,
	)
	if err != nil {
		return err
	}

	id, err = result.LastInsertId()
	if err != nil {
		return err
	}

	user.SetID(id)

	return nil
}

func (r *MySQLUserRepository) Update(ctx context.Context, user *domain.User) error {
	prefsJSON, err := json.Marshal(user.Preferences())
	if err != nil {
		return err
	}

	query := `
		UPDATE users
		SET
			username = ?,
			name = ?,
			preferences = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Username(),
		user.Name(),
		prefsJSON,
		user.ID(),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return apperr.New(
			apperr.CodeNotFound, "user_not_found", "The user to update was not found",
			apperr.WithOp("users.infrastructure.mysql_user_repository.Update"),
		)
	}

	return nil
}
