package Commands

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"proyecto1/DiskManagement"
	"proyecto1/Structs"
	"proyecto1/Utilities"
	"strings"
)

// Mount Función para montar particiones
func Mount(path string, name string) error {

	if path == "" || name == "" {
		return errors.New("path y Name son obligatorios")
	}

	fmt.Println("======Start MOUNT======")
	mountedPartitions := DiskManagement.GetMountedPartitions() // Obtener las particiones montadas
	file, err := Utilities.OpenFile(path)                      //abrimos el disco
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo en la ruta:", path)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error: No se pudo cerrar el archivo:", err)
		}
	}(file)

	var TempMBR Structs.MRB
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR desde el archivo")
		return err
	}

	var partition Structs.Partition
	var found = false
	// Convertir el nombre a comparar a un arreglo de bytes de longitud fija
	nameBytes := [16]byte{}
	copy(nameBytes[:], name)
	Structs.PrintMBR(TempMBR)
	for i := 0; i < 4; i++ {
		// Verificar si la partición es primaria y si el nombre coincide
		if bytes.Equal(TempMBR.Partitions[i].Name[:], nameBytes[:]) {

			if TempMBR.Partitions[i].Type[0] == 'e' {
				return errors.New("la partición es extendida")
			}

			if TempMBR.Partitions[i].Type[0] == 'l' {
				return errors.New("la partición no es de un tipo correcto")
			}

			if TempMBR.Partitions[i].Type[0] == 'p' {
				partition = TempMBR.Partitions[i]
				found = true
				fmt.Println("posicion de particion: ", i)
				break
			}
		}
	}

	if !found {
		return errors.New("No se encontró la partición con nombre: " + name)
	}
	// Verificar si la partición ya está montada
	if partition.Status[0] == '1' {
		return errors.New("la partición ya está montada")
	}
	//fmt.Printf("Partición encontrada: '%s' en posición %d\n", string(partition.Name[:]), partitionIndex+1)

	// Generar el ID de la partición
	diskID := generateDiskID(path)

	// Verificar si ya se ha montado alguna partición de este disco
	mountedPartitionsInDisk := mountedPartitions[diskID]

	// Verificar si la partición ya está montada
	for _, mountedPartition := range mountedPartitionsInDisk {
		if mountedPartition.Name == name {
			return errors.New("La partición' " + name + " ' ya está montada")
		}
	}

	var letter byte

	if len(mountedPartitionsInDisk) == 0 {
		// Es un nuevo disco, asignar la siguiente letra disponible
		if len(mountedPartitions) == 0 {
			letter = 'a'
		} else {
			lastDiskID := getLastDiskID()
			lastLetter := mountedPartitions[lastDiskID][0].ID[len(mountedPartitions[lastDiskID][0].ID)-1] // Obtener la última letra de la última partición montada en el último disco
			letter = lastLetter + 1                                                                       // Incrementar la letra
		}
	} else {
		// Utilizar la misma letra que las otras particiones montadas en el mismo disco
		letter = mountedPartitionsInDisk[0].ID[len(mountedPartitionsInDisk[0].ID)-1]
	}

	// Incrementar el número para esta partición
	carnet := "202203856" // Cambiar su carnet aquí
	lastTwoDigits := carnet[len(carnet)-2:]
	partitionID := fmt.Sprintf("%s%d%c", lastTwoDigits, len(mountedPartitions[diskID])+1, letter)

	// Actualizar el estado de la partición a montada y asignar el ID
	partition.Status[0] = '1'
	copy(partition.Id[:], partitionID)
	TempMBR.Partitions[len(mountedPartitions[diskID])] = partition
	mountedPartitions[diskID] = append(mountedPartitions[diskID], DiskManagement.MountedPartition{
		Path:   path,
		Name:   name,
		ID:     partitionID,
		Status: '1',
	})

	// Escribir el MBR actualizado al archivo
	if err := Utilities.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo sobrescribir el MBR en el archivo")
		return err
	}

	fmt.Printf("Partición montada con ID: %s\n", partitionID)

	fmt.Println("")
	// Imprimir el MBR actualizado
	fmt.Println("MBR actualizado:")
	Structs.PrintMBR(TempMBR)
	fmt.Println("")

	// Imprimir las particiones montadas (solo estan mientras dure la sesion de la consola)
	DiskManagement.PrintMountedPartitions()

	return nil
}

// Función para obtener el ID del último disco montado
func getLastDiskID() string {
	mountedPartitions := DiskManagement.GetMountedPartitions()
	var lastDiskID string
	for diskID := range mountedPartitions {
		lastDiskID = diskID
	}
	return lastDiskID
}

func generateDiskID(path string) string {
	return strings.ToLower(path)
}
