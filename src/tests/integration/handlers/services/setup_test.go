// src/tests/integration/handlers/setup_test.go
package testhandlers

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
)

var db *sqlx.DB

func TestMain(m *testing.M) {
	db = SetupTestDB()

	// Setup code: run your schemas here
	setupSchemas(db)

	// Run the tests
	code := m.Run()

	// if err := db.Close(); err != nil {
	// 	log.Fatal("Failed to close the database connection:", err)
	// }

	teardownSchemas(db)
	// Exit with the status code returned by the tests
	os.Exit(code)
}

func SetupTestDB() *sqlx.DB {
	// Load environment variables
	err := godotenv.Load("../../../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to the PostgreSQL test database
	dbUser := os.Getenv("DB_TEST_USER")
	dbPassword := os.Getenv("DB_TEST_PASSWORD")
	dbHost := os.Getenv("DB_TEST_HOST")
	dbPort := os.Getenv("DB_TEST_PORT")
	dbName := os.Getenv("DB_TEST_NAME")

	// Build the connection string
	dbConnectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := sqlx.Connect("postgres", dbConnectionString)
	if err != nil {
		log.Fatalf("Could not connect to the test database: %v", err)
	}

	return db
}

func setupSchemas(db *sqlx.DB) error {

	// Aqui você pode configurar o esquema de teste, por exemplo:
	statements := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		// Your CREATE TABLE statements here...
		`CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            first_name VARCHAR(255) NOT NULL,
            last_name VARCHAR(255) NOT NULL,
            username VARCHAR(255) NOT NULL,
            password VARCHAR(255),
            is_admin boolean
        );`,
		`CREATE TABLE clients (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			phone CHAR(12),
			is_mensal boolean,
			monthly_date DATE
		) INHERITS (users);`,
		`CREATE TABLE IF NOT EXISTS laundry_items (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			price numeric(10,2)
		);`,
		`CREATE TABLE IF NOT EXISTS laundry_services (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			status VARCHAR(15) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP,
			estimated_completion_date TIMESTAMP,
			total_price numeric(10,2),
			weight numeric(10,2),
			is_piece boolean,
			is_weight boolean,
			client_id UUID,
			is_paid boolean,
			FOREIGN KEY (client_id) REFERENCES clients(id)
		);`,
		`CREATE TABLE IF NOT EXISTS laundry_items_services (
			laundry_service_id UUID,
			laundry_item_id UUID,
			item_quantity INT,
			observation TEXT,
			PRIMARY KEY (laundry_service_id, laundry_item_id),
			FOREIGN KEY (laundry_service_id) REFERENCES laundry_services(id) ON DELETE CASCADE,
			FOREIGN KEY (laundry_item_id) REFERENCES laundry_items(id)
		);`,
		`CREATE TABLE address (
			address_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			street VARCHAR(255) NOT NULL,
			city VARCHAR(100) NOT NULL,
			state CHAR(2) NOT NULL,
			postal_code VARCHAR(20),
			number VARCHAR(5) NOT NULL,
			complement VARCHAR(255),
			landmark VARCHAR(255)
		);
		
		ALTER TABLE clients
		ADD COLUMN address_id UUID;
		
		ALTER TABLE clients
		ADD CONSTRAINT fk_address
		FOREIGN KEY (address_id) REFERENCES address(address_id)
		ON DELETE CASCADE;
		`,
	}

	for _, stmt := range statements {
		_, err := db.Exec(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func teardownSchemas(db *sqlx.DB) error {
	db.Exec("DELETE FROM laundry_items_services")
	db.Exec("DELETE FROM laundry_services")
	db.Exec("DELETE FROM address")
	db.Exec("DELETE FROM clients")
	db.Exec("DELETE FROM laundry_items")

	if err := db.Close(); err != nil {
		log.Fatal("Failed to close the database connection:", err)
	}

	return nil
}
