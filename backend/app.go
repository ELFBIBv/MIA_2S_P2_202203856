package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type CodeExecutionRequest struct {
	Code          string `json:"code"`
	FilePath      string `json:"filePath"`
	DiskPath      string `json:"diskPath"`
	PartitionName string `json:"partitionName"`
}

type CodeExecutionResponse struct {
	Output string `json:"output"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/execute", executeCode).Methods("POST")
	r.HandleFunc("/readmbr", ReadMBRHandler).Methods("POST")
	r.HandleFunc("/getPathDisks", GetPathMountedDisks).Methods("GET")
	r.HandleFunc("/getMountedPartitionsForPathDisk", GetMountedPartitionForPathDisk).Methods("POST")
	r.HandleFunc("/readfiles", ReadFilesHandler).Methods("POST")

	//para que acepte solo peticiones locales se usa:
	// c := cors.New(cors.Options{
	// 	...
	//	AllowedOrigins: []string{"http://localhost:3000", "http://[IP_ADDRESS]"},
	// })
	//esto es para que acepte peticiones desde cualquier lugar
	c := cors.New(cors.Options{
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		MaxAge:           300,
	})

	handler := c.Handler(r)

	log.Println("Servidor iniciado en el puerto 8080")
	// estos 2 de abajo son solamente para local
	// log.Fatal(http.ListenAndServe("localhost:8080", handler))
	// log.Fatal(http.ListenAndServe("127.0.0.1:8080", handler))
	//el de abajo ya acepta peticiones desde donde sea (0.0.0.0)
	log.Fatal(http.ListenAndServe(":8080", handler))
}
