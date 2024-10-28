package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"proyecto1/Analyzer"
	"proyecto1/DiskManagement"
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
func GetPathMountedDisks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//vamos a recolectar los paths no repetidos
	var paths []string
	disks := DiskManagement.GetMountedPartitions()
	for path, _ := range disks {
		paths = append(paths, path)
	}

	fmt.Println(paths)
	json.NewEncoder(w).Encode(paths)
}

func GetMountedPartitionForPathDisk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Estructura para la solicitud JSON
	var req struct {
		Path string `json:"path"`
	}

	// Decodificar el cuerpo JSON de la solicitud
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Error al procesar la solicitud", http.StatusBadRequest)
		return
	}

	// Validar que el campo `Path` no esté vacío
	if req.Path == "" {
		http.Error(w, "El parámetro 'path' es requerido", http.StatusBadRequest)
		return
	}

	// Obtener las particiones montadas
	partitions := DiskManagement.GetMountedPartitions()
	value, exist := partitions[req.Path]

	// Manejo de respuesta según existencia de la partición
	if exist {
		// Configurar el encabezado y enviar la respuesta JSON con las particiones
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(value)
	} else {
		_, nombreDisco := Utilities.GetParentDirectories(req.Path)
		http.Error(w, fmt.Sprintf("No se encontraron particiones montadas para el disco en la ruta '%s'", nombreDisco), http.StatusNotFound)
	}
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
