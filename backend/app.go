package main

import (
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
)

type CodeExecutionRequest struct {
	Code string `json:"code"`
}

type CodeExecutionResponse struct {
	Output string `json:"output"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/execute", executeCode).Methods("POST")
	r.HandleFunc("/readmbr", ReadMBRHandler).Methods("POST")

	c := cors.New(cors.Options{
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		MaxAge:           300,
	})

	handler := c.Handler(r)

	log.Println("Servidor iniciado en el puerto 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
