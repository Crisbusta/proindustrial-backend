package repository

import (
	"github.com/crisbusta/proindustrial-backend-public/internal/model"
	"github.com/jmoiron/sqlx"
)

type AuthRepo struct {
	db *sqlx.DB
}

func NewAuthRepo(db *sqlx.DB) *AuthRepo {
	return &AuthRepo{db: db}
}

func (r *AuthRepo) GetByEmail(email string) (*model.User, error) {
	var u model.User
	err := r.db.Get(&u, `SELECT * FROM users WHERE email = $1`, email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AuthRepo) GetByID(id string) (*model.User, error) {
	var u model.User
	err := r.db.Get(&u, `SELECT * FROM users WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

type CreateUserInput struct {
	Email              string
	PasswordHash       string
	CompanyID          *string
	Role               string
	MustChangePassword bool
}

func (r *AuthRepo) Create(in CreateUserInput) (*model.User, error) {
	var user model.User
	err := r.db.QueryRowx(`
		INSERT INTO users (email, password_hash, company_id, role, must_change_password)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *`,
		in.Email,
		in.PasswordHash,
		in.CompanyID,
		in.Role,
		in.MustChangePassword,
	).StructScan(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepo) ChangePassword(userID, passwordHash string) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET password_hash = $1, must_change_password = FALSE
		WHERE id = $2`,
		passwordHash,
		userID,
	)
	return err
}
