package Commands

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"proyecto1/DiskManagement"
	"proyecto1/FileSystem"
	"proyecto1/Structs"
	"proyecto1/Utilities"
	"strconv"
	"strings"
)

func Cat(paths []string, id string, permisos [3]byte) (error, string) {
	fmt.Println("======Start CAT======")
	fmt.Println("paths:", paths)

	// Verificar si el usuario ya está logueado buscando en las particiones montadas
	mountedPartitions := DiskManagement.GetMountedPartitions()
	var filepath string
	var partitionFound bool

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.ID == id { // Encuentra la partición correcta
				filepath = partition.Path
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		return errors.New("No se encontró ninguna partición montada con el ID '" + id + "' proporcionado"), "nil"
	}

	// Abrir archivo binario
	file, err := Utilities.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		return err, "nil"
	}

	var TempMBR Structs.MRB
	// Leer el MBR desde el archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return err, "nil"
	}

	var index = -1
	// Iterar sobre las particiones del MBR para encontrar la correcta
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), id) {
				if TempMBR.Partitions[i].Status[0] == '1' {
					index = i
				} else {
					fmt.Println("La partición no está montada")
					return errors.New("la partición no está montada"), "nil"
				}
				break
			}
		}
	}

	if index == -1 {
		fmt.Println("Partition not found")
		return errors.New("partición no encontrada"), "nil"
	}

	var tempSuperblock Structs.Superblock
	// Leer el Superblock desde el archivo binario
	if err := Utilities.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[index].Start)); err != nil {
		fmt.Println("Error: No se pudo leer el Superblock:", err)
		return err, "nil"
	}

	salida := ""
	// inode por path
	for _, path := range paths {
		// Buscar el archivo de usuarios /users.txt -> retorna índice del Inodo
		//indexInode := FileSystem.SearchInodeByPath("/users.txt", file, tempSuperblock)
		indexInode := FileSystem.SearchInodeByPath(path, file, tempSuperblock)

		posicion := int64(tempSuperblock.S_inode_start + indexInode*int32(binary.Size(Structs.Inode{})))

		var crrInode Structs.Inode
		// Leer el Inodo desde el archivo binario
		if err := Utilities.ReadObject(file, &crrInode, posicion); err != nil {
			fmt.Println("Error: No se pudo leer el Inodo:", err)
			return err, "nil"
		}

		permiso1, err := strconv.Atoi(string(crrInode.I_perm[:]))
		if err != nil {
			fmt.Println("Error: No se pudo convertir el permiso a entero:", err)
			return err, "nil"
		}

		permiso2, err := strconv.Atoi(string(permisos[:]))
		if err != nil {
			fmt.Println("Error: No se pudo convertir el permiso a entero", err)
		}

		if permiso1 > permiso2 {
			fmt.Println("No tiene permisos para leer el archivo" + path)
			salida += "No tiene permisos para leer el archivo" + path
			continue
		}

		Structs.PrintInode(crrInode)

		text, err := FileSystem.GetInodeFileData(crrInode, file, tempSuperblock)
		if err != nil {
			return err, "nil"
		}
		salida += text
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error: No se pudo cerrar el archivo:", err)
		}
	}(file)

	fmt.Println("======End CAT======")
	return nil, salida
}
