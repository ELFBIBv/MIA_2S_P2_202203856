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

// Función para verificar que un bloque del archivo esté lleno de ceros
func VerifyZeros(file *os.File, start int32, size int32) {
	zeros := make([]byte, size)
	_, err := file.ReadAt(zeros, int64(start))
	if err != nil {
		fmt.Println("Error al leer la sección eliminada:", err)
		return
	}

	// Verificar si todos los bytes leídos son ceros
	isZeroFilled := true
	for _, b := range zeros {
		if b != 0 {
			isZeroFilled = false
			break
		}
	}

	if isZeroFilled {
		fmt.Println("La partición eliminada está completamente llena de ceros.")
	} else {
		fmt.Println("Advertencia: La partición eliminada no está completamente llena de ceros.")
	}
}

// Función para llenar el espacio con ceros (\0)
func FillWithZeros(file *os.File, start int32, size int32) error {
	// Posiciona el archivo al inicio del área que debe ser llenada
	_, err := file.Seek(int64(start), 0)
	if err != nil {
		return err
	}

	// Crear un buffer lleno de ceros
	buffer := make([]byte, size)

	// Escribir los ceros en el archivo
	_, err = file.Write(buffer)
	if err != nil {
		fmt.Println("Error al llenar el espacio con ceros:", err)
		return err
	}

	fmt.Println("Espacio llenado con ceros desde el byte", start, "por", size, "bytes.")
	return nil
}
