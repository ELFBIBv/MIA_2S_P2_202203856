package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"proyecto1/Analyzer"
)

type CodeExecutionRequest struct {
	Code string `json:"code"`
}

type CodeExecutionResponse struct {
	Output string `json:"output"`
}

func executeCode(w http.ResponseWriter, r *http.Request) {
	var req CodeExecutionRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//App(req.Code)
	Analyzer.Analyze(req.Code)

	res := CodeExecutionResponse{Output: Analyzer.Salida}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/execute", executeCode).Methods("POST")

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
