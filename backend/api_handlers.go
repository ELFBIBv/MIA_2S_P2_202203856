package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"proyecto1/Analyzer"
	"proyecto1/Utilities"
)

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

type ReadMBRParams struct {
	Path string `json:"path"`
}

// Handler para leer el MBR y devolver las particiones
func ReadMBRHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var params ReadMBRParams
		// Decodificar el cuerpo JSON de la solicitud
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, "Error al procesar la solicitud", http.StatusBadRequest)
			return
		}
		// Validaciones
		if params.Path == "" {
			http.Error(w, "La ruta es requerida", http.StatusBadRequest)
			return
		}
		// Leer el MBR y obtener las particiones
		partitions, err := Utilities.ListPartitions(params.Path)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error al leer las particiones: %v", err), http.StatusInternalServerError)
			return
		}
		// Responder con las particiones en formato JSON
		json.NewEncoder(w).Encode(partitions)
	} else {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}
