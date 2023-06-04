package repositories

import (
	"database/sql"
	"log"
	models "metaedu-marketplace/models"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) InsertUser(user models.User) string {
	sqlStatement := `INSERT INTO users (
		name,
		email,
		photo,
		cover,
		verified,
		role,
		address,
		nonce,
		status,
		updated_at,
		created_at
	  ) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
	  )
	  RETURNING id`

	var id string

	err := r.db.QueryRow(sqlStatement, user.Name, user.Email, user.Photo, user.Cover, user.Verified, user.Role, user.Address, user.Nonce, user.Status, user.UpdatedAt, user.CreatedAt).Scan(&id)

	if err != nil {
		log.Fatalf("Query is cannot executed %v", err)
	}

	return id
}

func (r *UserRepository) GetUserByID(id uuid.UUID) (models.User, error) {
	sqlStatement := `SELECT id, name, email, photo, cover, role, address, nonce, updated_at, created_at FROM users where id = $1 LIMIT 1`

	var user models.User

	rows, err := r.db.Query(sqlStatement, id)

	if err != nil {
		return user, err
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&user.ID, &user.Name, &user.Email, &user.Photo, &user.Cover, &user.Role, &user.Address, &user.Nonce, &user.UpdatedAt, &user.CreatedAt)
		if err != nil {
			return user, err
		}
	}

	return user, err
}

func (r *UserRepository) GetSystemUser() (models.User, error) {
	sqlStatement := `SELECT id, name, email, photo, cover, role, address, nonce, updated_at, created_at 
					FROM users 
					WHERE role='admin' 
					LIMIT 1`

	var user models.User

	rows, err := r.db.Query(sqlStatement)

	if err != nil {
		return user, err
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&user.ID, &user.Name, &user.Email, &user.Photo, &user.Cover, &user.Role, &user.Address, &user.Nonce, &user.UpdatedAt, &user.CreatedAt)
		if err != nil {
			return user, err
		}
	}

	return user, err
}

func (r *UserRepository) GetUserByAddress(address string) models.User {
	sqlStatement := `SELECT id, name, email, photo, role, address, nonce, created_at, updated_at FROM users where address = $1 LIMIT 1`

	var user models.User

	rows, err := r.db.Query(sqlStatement, address)

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&user.ID, &user.Name, &user.Email, &user.Photo, &user.Role, &user.Address, &user.Nonce, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return user
		}
	}

	return user
}

func (r *UserRepository) UpdateUser(id uuid.UUID, user models.User) error {
	sqlStatement := `UPDATE users
	SET name = $2, email = $3, photo = $4, cover = $5, verified = $6, role = $7, address = $8, nonce = $9, status = $10, updated_at = $11
	WHERE id = $1;`

	_, err := r.db.Exec(sqlStatement, user.ID, user.Name, user.Email, user.Photo, user.Cover, user.Verified, user.Role, user.Address, user.Nonce, user.Status, user.UpdatedAt)

	if err != nil {
		log.Fatalf("Query is cannot executed %v", err)
	}

	return err
}
