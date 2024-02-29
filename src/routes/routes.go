package routes

import (
	clientshandlers "lavanderia/handlers/clients"
	itemshandlers "lavanderia/handlers/items"
	itemsserviceshandlers "lavanderia/handlers/laundryItemsServices"
	serviceshandlers "lavanderia/handlers/laundryServices"
	handlers "lavanderia/handlers/users"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// SetupRoutes configura as rotas HTTP para a aplicação.
func SetupRoutes(db *sqlx.DB) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/services/{serviceID}/items", itemsserviceshandlers.AddItemsServicesHandler(db)).Methods("POST")
	router.HandleFunc("/services/{serviceID}/items/{itemID}", itemsserviceshandlers.DeleteItemServiceHandler(db)).Methods("DELETE")
	router.HandleFunc("/services/{serviceID}/items/{itemID}", itemsserviceshandlers.UpdateItemServiceHandler(db)).Methods("PATCH")

	router.HandleFunc("/items", itemshandlers.CreateItemHandler(db)).Methods("POST")
	router.HandleFunc("/items", itemshandlers.ListItemsHandler(db)).Methods("GET")
	router.HandleFunc("/items/{id}", itemshandlers.ShowItemHandler(db)).Methods("GET")
	router.HandleFunc("/items/{id}", itemshandlers.DeleteItemHandler(db)).Methods("DELETE")
	router.HandleFunc("/items/{id}", itemshandlers.UpdateItemHandler(db)).Methods("PUT")

	router.HandleFunc("/clients", clientshandlers.CreateClientHandler(db)).Methods("POST")
	router.HandleFunc("/clients", clientshandlers.ListClientsHandler(db)).Methods("GET")
	router.HandleFunc("/clients/{id}", clientshandlers.ShowClientHandler(db)).Methods("GET")
	router.HandleFunc("/clients/{id}", clientshandlers.DeleteClientHandler(db)).Methods("DELETE")
	router.HandleFunc("/clients/{id}", clientshandlers.UpdateClientHandler(db)).Methods("PUT")

	router.HandleFunc("/services", serviceshandlers.CreateServicesHandler(db)).Methods("POST")
	router.HandleFunc("/services", serviceshandlers.ListServicesHandler(db)).Methods("GET")
	router.HandleFunc("/services/{id}", serviceshandlers.ShowServiceHandler(db)).Methods("GET")
	router.HandleFunc("/services/{id}", serviceshandlers.UpdateServiceHandler(db)).Methods("PUT")
	router.HandleFunc("/services/{id}", serviceshandlers.DeleteServiceHandler(db)).Methods("DELETE")

	router.HandleFunc("/users", handlers.CreateUserHandler(db)).Methods("POST")
	router.HandleFunc("/users", handlers.ListUsersHandler(db)).Methods("GET")
	router.HandleFunc("/users/{id}", handlers.UpdateUserHandler(db)).Methods("PATCH")
	router.HandleFunc("/users/{id}", handlers.DeleteUserHandler(db)).Methods("DELETE")

	return router
}
