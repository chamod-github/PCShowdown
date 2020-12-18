package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Server struct to store server components
type Server struct {
	router *mux.Router
	db     *Database
}

// createPC creates a PC and writes back new PC info
func (server *Server) createPC(writer http.ResponseWriter, request *http.Request) {
	var pc PC

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&pc)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	// Add to the db
	links, err := server.db.addPC(pc)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	// Encode the new links as json
	writer.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(writer)
	err = encoder.Encode(links)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getPCs gets all PCs
func (server *Server) getPCS(writer http.ResponseWriter, request *http.Request) {
	oID, err := strconv.Atoi(mux.Vars(request)["page_number"])
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	limit, err := strconv.Atoi(mux.Vars(request)["limit"])
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	pcs, err := server.db.getPCS(oID, limit)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(writer)
	err = encoder.Encode(pcs)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

}

// getPCs gets all PCs
func (server *Server) getPC(writer http.ResponseWriter, request *http.Request) {
	linkID := mux.Vars(request)["link_id"]

	pc, err := server.db.getPC(linkID)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(writer)
	err = encoder.Encode(pc)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

// setupRoutes sets up routes for the router
func (server *Server) setupRoutes() {
	if server.router == nil {
		log.Fatalln("Router is not initialized.")
	}

	// API

	server.router.StrictSlash(true)
	subrouter := server.router.PathPrefix("/api/v1").Subrouter()
	// PCs
	subrouter.HandleFunc("/pcs", server.createPC).Methods(http.MethodPost)
	// page number and how many per page
	subrouter.HandleFunc("/pcs/{page_number}/{limit}", server.getPCS).Methods(http.MethodGet)
	subrouter.HandleFunc("/pcs/{link_id}", server.getPC).Methods(http.MethodGet)

	// Website

	// server.router.HandleFunc("/", getHomePage).Methods(http.MethodGet)
	server.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./dist/static")))
	server.router.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./dist/static"))))

}

// initializeServer initializes server components
func (server *Server) initializeServer() {
	server.router = mux.NewRouter()
	server.setupRoutes()
}

// StartServer starts the server
func (server *Server) StartServer(addr string) {
	// Setup routes and initialize
	server.initializeServer()
	// Start the server
	log.Fatal(http.ListenAndServe(addr, server.router))
}
