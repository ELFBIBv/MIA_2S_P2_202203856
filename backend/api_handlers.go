package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"proyecto1/Analyzer"
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
		partitions, err := ListPartitions(params.Path)
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

// vamos a retornar el path de todos los discos guardados en analizer
func GetPathDisks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Analyzer.PathDisks)
}

func ReadFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req CodeExecutionRequest

		// Decodificar el cuerpo JSON de la solicitud
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Error al procesar la solicitud", http.StatusBadRequest)
			return
		}

		// Validaciones
		if req.FilePath == "" || req.DiskPath == "" || req.PartitionName == "" {
			http.Error(w, "Todos los parámetros son requeridos", http.StatusBadRequest)
			return
		}

		// Llamar a recolectFiles para obtener los archivos
		erro, arrayFiles := recolectFiles(req.FilePath, req.DiskPath, req.PartitionName)
		if erro != nil {
			http.Error(w, fmt.Sprintf("Error al recolectar los archivos: %v", erro), http.StatusInternalServerError)
			return
		}

		// Responder con el array de archivos en formato JSON
		json.NewEncoder(w).Encode(arrayFiles)
	} else {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}
