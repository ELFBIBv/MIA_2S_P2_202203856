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

func Rmusr(name string, id string) error {
	fmt.Println("======INICIO RMUSR======")
	fmt.Println("Name:", name)

	// Verificar si el usuario ya está logueado buscando en las particiones montadas
	mountedPartitions := DiskManagement.GetMountedPartitions()
	filepath := ""
	partitionFound := false

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if strings.TrimSpace(partition.ID) == strings.TrimSpace(id) { // Encuentra la partición correcta
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
		return fmt.Errorf("no se encontró ninguna partición montada con el ID '%s' proporcionado(1)", id)
	}

	// Abrir archivo binario
	file, err := Utilities.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		return err
	}

	var TempMBR Structs.MRB
	// Leer el MBR desde el archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return err
	}

	var index = -1
	// Iterar sobre las particiones del MBR para encontrar la correcta
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			fmt.Println("TempMBR.Partitions[i].Id[:]: ", string(TempMBR.Partitions[i].Id[:]))
			fmt.Println("id: ", id)
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), id) {
				if TempMBR.Partitions[i].Status[0] == '1' {
					index = i
				} else {
					fmt.Println("La partición no está montada")
					return errors.New("la partición no está montada")
				}
				break
			}
		}
	}

	if index == -1 {
		return errors.New("No se encontró ninguna partición con el ID '" + id + "' proporcionado(2)")
	}

	// Imprimir información de la partición
	//Structs.PrintPartition(TempMBR.Partitions[index])

	var tempSuperblock Structs.Superblock
	// Leer el Superblock desde el archivo binario

	if err := Utilities.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[index].Start)); err != nil {
		fmt.Println("Error: No se pudo leer el Superblock:", err)
		return err
	}

	// Buscar el archivo de usuarios /users.txt -> retorna índice del Inodo
	indexInode := FileSystem.SearchInodeByPath("/users.txt", file, tempSuperblock)

	posicion := int64(tempSuperblock.S_inode_start + indexInode*int32(binary.Size(Structs.Inode{})))

	var crrInode Structs.Inode
	// Leer el Inodo desde el archivo binario
	if err := Utilities.ReadObject(file, &crrInode, posicion); err != nil {
		fmt.Println("Error: No se pudo leer el Inodo:", err)
		return err
	}

	// Leer datos del archivo
	data, err := FileSystem.GetInodeFileData(crrInode, file, tempSuperblock)
	if err != nil {
		fmt.Println("Error: No se pudo leer el archivo:", err)
		return err
	}

	// Dividir la cadena en líneas
	lines := strings.Split(data, "\n")

	//revisamos que el grupo no este desactivado
	for _, line := range lines {
		palabras := strings.Split(line, ",")
		if strings.Contains(palabras[0], "0") && strings.Contains(palabras[1], "G") && strings.Contains(palabras[2], name) {
			fmt.Println("Error: El usuario ya está desactivado")
			return errors.New("el usuario ya está desactivado")
		}
	}

	found := false
	//revisamos que el usuario exista
	for _, line := range lines {
		palabras := strings.Split(line, ",")
		if palabras[1] == "U" && palabras[3] == name {
			found = true
		}
	}

	if !found {
		fmt.Println("Error: El usuario no existe")
		return errors.New("el usuario no existe")
	}

	possition := 0

	// Iterar a través de las líneas para verificar las credenciales
	for _, line := range lines {
		fmt.Println("linea: ", line)
		palabras := strings.Split(line, ",")
		if palabras[1] == "G" {
			//sumamos la cantidad de caracteres de la linea y el salto de linea
			possition += len(line) + 1
			continue
		}

		if len(palabras) == 5 {
			if palabras[3] != name {
				//sumamos la cantidad de caracteres de la linea y el salto de linea
				possition += len(line) + 1
				continue
			}
		}

		if palabras[3] == name {
			break
		}
	}

	//vamos a ver en cual file esta
	filePoss := int64(possition) / int64(tempSuperblock.S_block_size)

	//vamos a obtener el bloque
	var crrFileBlock Structs.Fileblock
	offset := int64(tempSuperblock.S_block_start + crrInode.I_block[filePoss]*int32(binary.Size(Structs.Fileblock{})))
	if err := Utilities.ReadObject(file, &crrFileBlock, offset); err != nil {
		return err
	}
	//fmt.Println("antes")
	//Structs.PrintFileblock(crrFileBlock)
	//fmt.Println("fileblock", crrFileBlock.B_content)

	//obtenemos la posicion relativa del bit a borrar
	//la posicion relativa es la posicion en el bloque
	//menos la cantidad de caracteres de las lineas anteriores
	relativePosition := int64(possition) % int64(tempSuperblock.S_block_size)
	crrFileBlock.B_content[relativePosition] = byte(48)

	//fmt.Println("despues")
	//Structs.PrintFileblock(crrFileBlock)

	//reescribimos el bloque
	if err := Utilities.WriteObject(file, crrFileBlock, offset); err != nil {
		return err
	}

	fmt.Println("======FIN RMUSR======")

	return nil
}
