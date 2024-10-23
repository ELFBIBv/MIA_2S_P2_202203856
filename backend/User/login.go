package User

import (
	"encoding/binary"
	"errors"
	"fmt"
	"proyecto1/DiskManagement"
	"proyecto1/FileSystem"
	"proyecto1/Structs"
	"proyecto1/Utilities"
	"strings"
)

func Login(user string, pass string, id string) (error, [3]byte) {
	fmt.Println("======Start LOGIN======")
	fmt.Println("User:", user)
	fmt.Println("Pass:", pass)
	fmt.Println("Id:", id)

	// Verificar si el usuario ya está logueado buscando en las particiones montadas
	mountedPartitions := DiskManagement.GetMountedPartitions()
	var filepath string
	var partitionFound bool
	var login = false

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.ID == id && partition.LoggedIn { // Verifica si ya está logueado
				fmt.Println("Ya existe un usuario logueado!")
				return errors.New("ya existe un usuario logueado"), [3]byte{'0', '0', '0'}
			}
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
		return errors.New("No se encontró ninguna partición montada con el ID '" + id + "' proporcionado"), [3]byte{'0', '0', '0'}
	}

	// Abrir archivo binario
	file, err := Utilities.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		return err, [3]byte{'0', '0', '0'}
	}

	var TempMBR Structs.MRB
	// Leer el MBR desde el archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return err, [3]byte{'0', '0', '0'}
	}

	// Imprimir el MBR
	//Structs.PrintMBR(TempMBR)
	//fmt.Println("-------------")

	var index = -1
	// Iterar sobre las particiones del MBR para encontrar la correcta
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), id) {
				if TempMBR.Partitions[i].Status[0] == '1' {
					index = i
				} else {
					fmt.Println("La partición no está montada")
					return errors.New("la partición no está montada"), [3]byte{'0', '0', '0'}
				}
				break
			}
		}
	}

	if index == -1 {
		fmt.Println("Partition not found")
		return errors.New("partición no encontrada"), [3]byte{'0', '0', '0'}
	}

	var tempSuperblock Structs.Superblock
	// Leer el Superblock desde el archivo binario
	if err := Utilities.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[index].Start)); err != nil {
		fmt.Println("Error: No se pudo leer el Superblock:", err)
		return err, [3]byte{'0', '0', '0'}
	}

	// Buscar el archivo de usuarios /users.txt -> retorna índice del Inodo
	indexInode := FileSystem.SearchInodeByPath("/users.txt", file, tempSuperblock)

	posicion := int64(tempSuperblock.S_inode_start + indexInode*int32(binary.Size(Structs.Inode{})))

	var crrInode Structs.Inode
	// Leer el Inodo desde el archivo binario
	if err := Utilities.ReadObject(file, &crrInode, posicion); err != nil {
		fmt.Println("Error: No se pudo leer el Inodo:", err)
		return err, [3]byte{'0', '0', '0'}
	}

	// Leer datos del archivo
	data, err := FileSystem.GetInodeFileData(crrInode, file, tempSuperblock)
	if err != nil {
		fmt.Println("Error: No se pudo leer el archivo:", err)
		return err, [3]byte{'0', '0', '0'}
	}

	// Dividir la cadena en líneas
	lines := strings.Split(data, "\n")

	permisos := [3]byte{0, 0, 0}
	// Iterar a través de las líneas para verificar las credenciales
	for _, line := range lines {
		words := strings.Split(line, ",")

		if len(words) == 5 {
			if words[3] == user && words[4] == pass {
				login = true
				if words[3] == "root" {
					permisos = [3]byte{'7', '7', '7'}
				} else {
					permisos = [3]byte{'6', '6', '4'}
				}
				break
			}
		}
	}

	// Imprimir información del Inodo
	//fmt.Println("Inode", crrInode.I_block)

	// Si las credenciales son correctas y marcamos como logueado
	if login {
		fmt.Println("Usuario logueado con exito")
		DiskManagement.MarkPartitionAsLoggedIn(id) // Marcar la partición como logueada
	}

	fmt.Println("Permisos:", permisos)

	fmt.Println("======End LOGIN======")
	return nil, permisos
}
