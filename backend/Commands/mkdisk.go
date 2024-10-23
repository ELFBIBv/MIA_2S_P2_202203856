package Commands

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"proyecto1/Structs"
	"proyecto1/Utilities"
	"time"
)

func Mkdisk(size int, fit string, unit string, path string) error {

	// Validar los parametros
	err := validatemkDisk(size, path, unit, fit)
	if err != nil {
		return err
	}

	fmt.Println("======INICIO MKDISK======")
	fmt.Println("Size:", size)
	fmt.Println("Fit:", fit)
	fmt.Println("Unit:", unit)
	fmt.Println("Path:", path)

	// Create file
	err = Utilities.CreateFile(path)
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	//Si el usuario especifica unit = "k" (Kilobytes), el tamaño se multiplica por 1024 para convertirlo a bytes.
	if unit == "k" {
		size = size * 1024
	} else {
		//Si el usuario especifica unit = "m" (Megabytes), el tamaño se multiplica por 1024 * 1024 para convertirlo a MEGA bytes.
		size = size * 1024 * 1024
	}

	// Open bin file
	file, err := Utilities.OpenFile(path)
	if err != nil {
		return err
	}

	// Escribir los 0 en el archivo
	for i := 0; i < size; i++ {
		err := Utilities.WriteObject(file, byte('0'), int64(i))
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}

	// Crear MRB
	var newMRB Structs.MRB
	newMRB.MbrSize = int32(size)
	newMRB.Signature = rand.Int31() // Numero random rand.Int31() genera solo números no negativos
	copy(newMRB.Fit[:], fit)

	// Obtener la fecha del sistema en formato YYYY-MM-DD
	currentTime := time.Now()
	formattedDate := currentTime.Format("2006-01-02 15:04")
	copy(newMRB.CreationDate[:], formattedDate)

	// Escribir el archivo
	if err := Utilities.WriteObject(file, newMRB, 0); err != nil {
		return err
	}

	//ahora vamos a imprimir el MBR
	var TempMBR Structs.MRB
	// Leer el archivo
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		return err
	}

	// Print object
	Structs.PrintMBR(TempMBR)

	// Cerrar el archivo
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	fmt.Println("======FIN MKDISK======")
	return nil
}

func validatemkDisk(size int, path string, unit string, fit string) error {
	// Validar fit bf/ff/wf
	if fit != "bf" && fit != "wf" && fit != "ff" {
		fmt.Println("Error: Fit debe ser bf, wf or ff")
		return errors.New("fit debe ser bf, wf or ff")
	}

	// Validar size > 0
	if size <= 0 {
		fmt.Println("Error: Size debe ser mayor a  0")
		return errors.New("Size debe ser mayor a  0")
	}

	// Validar unidar k - m
	if unit != "k" && unit != "m" {
		fmt.Println("Error: Las unidades validas son k o m")
		return errors.New("las unidades validas son k o m")
	}

	if path == "" {
		fmt.Println("Error: Path is required")
		return errors.New("path is required")
	}

	return nil
}
