package data

import (
	"context"
	"database/sql"
	"log"
	"time"
)

const dbTimeout = time.Second * 3

var db *sql.DB

// New is the function used to create an instance of the data package. It returns the type
// Model, which embeds all the types we want to be available to our application.
func New(dbPool *sql.DB) Models {
	db = dbPool

	return Models{
		Company: Company{},
	}
}

// Models is the type for this package. Note that any model that is included as a member
// in this type is available to us throughout the application, anywhere that the
// app variable is used, provided that the model is also added in the New function.
type Models struct {
	Company Company
}

// Company is the structure which holds one company from the database.
type Company struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	NumOfEmployees int       `json:"num_of_employees"`
	Registered     bool      `json:"registered"`
	Type           string    `json:"type"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// GetAll returns a slice of all companies, sorted by last name
func (u *Company) GetAll() ([]*Company, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, name, description, num_of_employees, registered, type, created_at, updated_at
	from companies order by name`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []*Company

	for rows.Next() {
		var company Company
		err := rows.Scan(
			&company.ID,
			&company.Name,
			&company.Description,
			&company.NumOfEmployees,
			&company.Registered,
			&company.Type,
			&company.CreatedAt,
			&company.UpdatedAt,
		)
		if err != nil {
			log.Println("Error scanning", err)
			return nil, err
		}

		companies = append(companies, &company)
	}

	return companies, nil
}

// GetByEmail returns one company by email
func (u *Company) GetByEmail(email string) (*Company, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, name, description, num_of_employees, registered, type, created_at, updated_at from companies where email = $1`

	var company Company
	row := db.QueryRowContext(ctx, query, email)

	err := row.Scan(
		&company.ID,
		&company.Name,
		&company.Description,
		&company.NumOfEmployees,
		&company.Registered,
		&company.Type,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &company, nil
}

// GetOne returns one company by id
func (u *Company) GetOne(id int) (*Company, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, name, description, num_of_employees, registered, type, created_at, updated_at from companies where id = $1`

	var company Company
	row := db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&company.ID,
		&company.Name,
		&company.Description,
		&company.NumOfEmployees,
		&company.Registered,
		&company.Type,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &company, nil
}

// Update updates one company in the database, using the information
// stored in the receiver u
func (u *Company) Update() error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `update companies set
		email = $1,
		first_name = $2,
		last_name = $3,
		user_active = $4,
		updated_at = $5
		where id = $6
	`

	_, err := db.ExecContext(ctx, stmt,
		u.Name,
		u.Description,
		u.NumOfEmployees,
		u.Type,
		time.Now(),
		u.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// Delete deletes one company from the database, by Company.ID
func (u *Company) Delete() error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `delete from companies where id = $1`

	_, err := db.ExecContext(ctx, stmt, u.ID)
	if err != nil {
		return err
	}

	return nil
}

// DeleteByID deletes one company from the database, by ID
func (u *Company) DeleteByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `delete from companies where id = $1`

	_, err := db.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}

// Insert inserts a new company into the database, and returns the ID of the newly inserted row
func (u *Company) Insert(company Company) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var newID int
	stmt := `insert into companies (email, name, description, num_of_employees, registered, type, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7) returning id`

	err := db.QueryRowContext(ctx, stmt,
		company.Name,
		company.Description,
		company.NumOfEmployees,
		company.Registered,
		company.Type,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}
