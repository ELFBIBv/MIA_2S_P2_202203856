package Utilities

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CreateFile Funcion para crear un archivo binario
func CreateFile(name string) error {
	//Se asegura que el archivo existe
	dir := filepath.Dir(name)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Println("Err CreateFile dir==", err)
		return err
	}

	// Crear archivo
	if _, err := os.Stat(name); os.IsNotExist(err) {
		file, err := os.Create(name)
		if err != nil {
			fmt.Println("Err CreateFile create==", err)
			return err
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				fmt.Println("Err CreateFile close==", err)
			}
		}(file)
	}
	return nil
}

// OpenFile Funcion para abrir un archivo binario ead/write mode
func OpenFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Err OpenFile==", err)
		return nil, err
	}
	return file, nil
}

// DeleteFile Funcion para eliminar un archivo
func DeleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		fmt.Println("Err DeleteFile==", err)
		return err
	}
	return nil
}

// WriteObject Funcion para escribir un objecto en un archivo binario
func WriteObject(file *os.File, data interface{}, position int64) error {
	_, err := file.Seek(position, 0)
	if err != nil {
		fmt.Println("Err WriteObject==", err)
	}
	err = binary.Write(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Err WriteObject==", err)
		return err
	}
	return nil
}

// ReadObject Funcion para leer un objeto de un archivo binario
func ReadObject(file *os.File, data interface{}, position int64) error {
	_, err := file.Seek(position, 0)
	if err != nil {
		return err
	}
	err = binary.Read(file, binary.LittleEndian, data)
	if err != nil {
		return err
	}
	return nil
}

func GetParentDirectories(path string) ([]string, string) {
	// Normalizar el path
	path = filepath.Clean(path)

	// Dividir el path en sus componentes
	components := strings.Split(path, string(filepath.Separator))

	// Lista para almacenar las rutas de las carpetas padres
	var parentDirs []string

	// Construir las rutas de las carpetas padres, excluyendo la última carpeta
	for i := 1; i < len(components)-1; i++ {
		parentDirs = append(parentDirs, components[i])
	}

	// La última carpeta es la carpeta de destino
	destDir := components[len(components)-1]

	return parentDirs, destDir
}

func SplitStringIntoChunks(s string) []string {
	var chunks []string
	for i := 0; i < len(s); i += 64 {
		end := i + 64
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}
