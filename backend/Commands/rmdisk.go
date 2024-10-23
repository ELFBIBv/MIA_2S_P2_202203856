package Commands

import (
	"errors"
	"fmt"
	"proyecto1/Utilities"
)

func Rmdisk(path string) error {
	fmt.Println("======Start RMDISK======")
	fmt.Println("Path:", path)

	// Validar path
	if path == "" {
		return errors.New("path is required")
	}

	// Eliminar archivo
	err := Utilities.DeleteFile(path)
	if err != nil {
		return err
	}
	fmt.Println("=====Disk deleted successfully=====")
	return nil
}
