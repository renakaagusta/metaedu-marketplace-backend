package repositories

import (
	"database/sql"
	"fmt"
	"log"
	models "metaedu-marketplace/models"
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
		verified,
		role,
		address,
		nonce,
		updated_at
	  ) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8
	  )
	  RETURNING id`

	var id string

	err := r.db.QueryRow(sqlStatement, user.Name, user.Email, user.Photo, user.Verified, user.Role, user.Address, user.Nonce, user.UpdatedAt).Scan(&id)

	if err != nil {
		log.Fatalf("Query is cannot executed %v", err)
	}

	return id
}

func (r *UserRepository) GetUserByAddress(address string) models.User {
	sqlStatement := `SELECT id, name, email, photo, role, address, nonce FROM users where address = $1 LIMIT 1`

	var user models.User

	rows, err := r.db.Query(sqlStatement, address)

	fmt.Println(err)

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&user.ID, &user.Name, &user.Email, &user.Photo, &user.Role, &user.Address, &user.Nonce)
		if err != nil {
			return user
		}
	}

	return user
}

func (r *UserRepository) UpdateUser(user models.User) error {
	sqlStatement := `UPDATE users
	SET name = $2, email = $3, photo = $4, verified = $5, role = $6, address = $7, nonce = $8, updated_at = $9
	WHERE id = $1;`

	_, err := r.db.Exec(sqlStatement, user.ID, user.Name, user.Email, user.Photo, user.Verified, user.Role, user.Address, user.Nonce, user.UpdatedAt)

	if err != nil {
		log.Fatalf("Query is cannot executed %v", err)
	}

	return err
}
