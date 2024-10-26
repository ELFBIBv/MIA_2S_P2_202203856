package Utilities

import (
	"fmt"
	"os"
	"proyecto1/Structs"
	"strings"
)

// Estructura para representar una partición en JSON
type PartitionInfo struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Start  int32  `json:"start"`
	Size   int32  `json:"size"`
	Status string `json:"status"`
}

func ListPartitions(path string) ([]PartitionInfo, error) {
	// Abrir el archivo binario
	file, err := OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Error al cerrar el archivo: %v\n", err)
		}
	}(file)
	// Crear una variable para almacenar el MBR
	var mbr Structs.MRB
	// Leer el MBR desde el archivo
	err = ReadObject(file, &mbr, 0) // Leer desde la posición 0
	if err != nil {
		return nil, fmt.Errorf("Error al leer el MBR: %v", err)
	}
	// Crear una lista de particiones basada en el MBR
	var partitions []PartitionInfo
	for _, partition := range mbr.Partitions {
		if partition.Size > 0 { // Solo agregar si la partición tiene un tamaño
			// Limpiar el nombre para eliminar caracteres nulos
			partitionName := strings.TrimRight(string(partition.Name[:]), "\x00")
			partitions = append(partitions, PartitionInfo{
				Name:   partitionName,
				Type:   strings.TrimRight(string(partition.Type[:]), "\x00"),
				Start:  partition.Start,
				Size:   partition.Size,
				Status: strings.TrimRight(string(partition.Status[:]), "\x00"),
			})
		}
	}
	return partitions, nil
}
