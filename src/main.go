package main

import (
	"fmt"
	router "lavanderia/routes"
	"log"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

func main() {
	// Carrega variáveis de ambiente do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env")
	}

	// Connect to your PostgreSQL database
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Build connection string
	dbConnectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sqlx.Connect("postgres", dbConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	routes := router.SetupRoutes(db)

	env := os.Getenv("ENV") // 'development' or 'production'
	appURL := os.Getenv("FRONTEND_URL")
	appPort := os.Getenv("FRONTEND_PORT")

	var address string
	if env == "development" {
		address = fmt.Sprintf("%s:%s", appURL, appPort)
	} else { // In production, assume APP_URL includes any necessary port information or is routed correctly
		address = appURL
	}

	// Configurar o CORS com as opções desejadas
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{address}, // Replace "*" with your Next.js app origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true, // Important: this allows cookies and authorization headers
	})

	// Use o middleware CORS para envolver suas rotas
	handler := c.Handler(routes)

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", handler)
}
