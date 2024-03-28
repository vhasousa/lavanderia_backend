package routes

import (
	clientshandlers "lavanderia/handlers/clients"
	itemshandlers "lavanderia/handlers/items"
	itemsserviceshandlers "lavanderia/handlers/laundryItemsServices"
	serviceshandlers "lavanderia/handlers/laundryServices"
	handlers "lavanderia/handlers/users"
	middleware "lavanderia/middlewares"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// SetupRoutes configura as rotas HTTP para a aplicação.
func SetupRoutes(db *sqlx.DB) *mux.Router {
	router := mux.NewRouter()

	// Wrap the routes you want to protect with JWTAuthentication middleware
	protectedRoutes := router.PathPrefix("").Subrouter()
	protectedRoutes.Use(middleware.JWTAuthentication)

	protectedRoutes.Handle("/services/{serviceID}/items", middleware.RoleAuthorization("Admin")(http.HandlerFunc(itemsserviceshandlers.AddItemsServicesHandler(db)))).Methods("POST")
	protectedRoutes.Handle("/services/{serviceID}/items/{itemID}", middleware.RoleAuthorization("Admin")(http.HandlerFunc(itemsserviceshandlers.DeleteItemServiceHandler(db)))).Methods("DELETE")
	protectedRoutes.Handle("/services/{serviceID}/items/{itemID}", middleware.RoleAuthorization("Admin")(http.HandlerFunc(itemsserviceshandlers.UpdateItemServiceHandler(db)))).Methods("PATCH")

	protectedRoutes.Handle("/items", middleware.RoleAuthorization("Admin")(http.HandlerFunc(itemshandlers.CreateItemHandler(db)))).Methods("POST")
	router.HandleFunc("/items", itemshandlers.ListItemsHandler(db)).Methods("GET")
	router.HandleFunc("/items/{id}", itemshandlers.ShowItemHandler(db)).Methods("GET")
	protectedRoutes.Handle("/items/{id}", middleware.RoleAuthorization("Admin")(http.HandlerFunc(itemshandlers.DeleteItemHandler(db)))).Methods("DELETE")
	protectedRoutes.Handle("/items/{id}", middleware.RoleAuthorization("Admin")(http.HandlerFunc(itemshandlers.UpdateItemHandler(db)))).Methods("PUT")

	protectedRoutes.Handle("/clients", middleware.RoleAuthorization("Admin")(http.Handler(clientshandlers.CreateClientHandler(db)))).Methods("POST")
	protectedRoutes.Handle("/clients", middleware.RoleAuthorization("Admin")(http.HandlerFunc(clientshandlers.ListClientsHandler(db)))).Methods("GET")
	protectedRoutes.Handle("/clients/{id}", middleware.RoleAuthorization("Admin")(http.HandlerFunc(clientshandlers.ShowClientHandler(db)))).Methods("GET")
	protectedRoutes.Handle("/clients/{id}", middleware.RoleAuthorization("Admin")(http.HandlerFunc(clientshandlers.DeleteClientHandler(db)))).Methods("DELETE")
	protectedRoutes.Handle("/clients/{id}", middleware.RoleAuthorization("Admin")(http.HandlerFunc(clientshandlers.UpdateClientHandler(db)))).Methods("PUT")
	protectedRoutes.Handle("/clients/{id}/renew", middleware.RoleAuthorization("Admin")(http.HandlerFunc(clientshandlers.RenewMonthlyFeeHandler(db)))).Methods("PATCH")

	protectedRoutes.Handle("/services", middleware.RoleAuthorization("Admin")(http.HandlerFunc(serviceshandlers.CreateServicesHandler(db)))).Methods("POST")
	protectedRoutes.Handle("/services", middleware.RoleAuthorization("Admin")(http.HandlerFunc(serviceshandlers.ListServicesHandler(db)))).Methods("GET")
	protectedRoutes.Handle("/services/client/{id}", middleware.RoleAuthorization("Admin", "Client")(http.HandlerFunc(serviceshandlers.ListServicesByClientHandler(db)))).Methods("GET")
	protectedRoutes.Handle("/services/{id}", middleware.RoleAuthorization("Admin", "Client")(http.HandlerFunc(serviceshandlers.ShowServiceHandler(db)))).Methods("GET")
	protectedRoutes.Handle("/services/{id}", middleware.RoleAuthorization("Admin")(http.HandlerFunc(serviceshandlers.UpdateServiceHandler(db)))).Methods("PUT")
	protectedRoutes.Handle("/services/{id}", middleware.RoleAuthorization("Admin")(http.HandlerFunc(serviceshandlers.DeleteServiceHandler(db)))).Methods("DELETE")

	router.HandleFunc("/login", handlers.LoginHandler(db)).Methods("POST")
	protectedRoutes.Handle("/users", middleware.RoleAuthorization("Admin")(http.HandlerFunc(handlers.CreateUserHandler(db)))).Methods("POST")
	protectedRoutes.Handle("/users", middleware.RoleAuthorization("Admin")(http.HandlerFunc(handlers.ListUsersHandler(db)))).Methods("GET")
	protectedRoutes.Handle("/users/{id}", middleware.RoleAuthorization("Admin")(http.HandlerFunc(handlers.UpdateUserHandler(db)))).Methods("PATCH")
	protectedRoutes.Handle("/users/{id}", middleware.RoleAuthorization("Admin")(http.HandlerFunc(handlers.DeleteUserHandler(db)))).Methods("DELETE")
	router.HandleFunc("/api/auth/status", handlers.AuthStatusHandler)
	router.HandleFunc("/api/auth/logout", handlers.LogoutHandler)

	return router
}
