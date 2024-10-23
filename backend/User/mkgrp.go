package User

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

func Mkgrp(name string, id string) error {
	fmt.Println("======INICIO MKGRP======")
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

	//revisamos que el grupo no este activo
	for _, line := range lines {
		palabras := strings.Split(line, ",")
		if !strings.Contains(palabras[0], "0") && strings.Contains(palabras[1], "G") && strings.Contains(palabras[2], name) {
			fmt.Println("Error: El grupo ya está activo")
			return errors.New("el grupo ya está activo")
		}
	}

	cantidadG := 1

	// Iterar a través de las líneas para verificar las credenciales
	for {
		found := false
		for _, line := range lines {
			palabras := strings.Split(line, ",")
			if palabras[0] == strconv.Itoa(cantidadG) && palabras[1] == "G" {
				found = true
				break
			}
		}
		if !found {
			break
		}
		cantidadG++
	}

	if cantidadG > 9 {
		return errors.New("no se pueden tener más de 9 grupos")
	}

	for _, line := range lines {
		palabras := strings.Split(line, ",")
		if palabras[0] == "0" && palabras[1] == "G" && palabras[2] == name {
			err := activarGrupoDesactivado(file, lines, name, tempSuperblock, crrInode, cantidadG)
			if err != nil {
				return err
			}
			fmt.Println("======FIN MKGRP======")
			return nil
		}

	}

	// Crear nuevo grupo
	text := fmt.Sprintf("%d,G,%s", cantidadG, name)

	// Escribir en el archivo
	if err := FileSystem.AppendToFileBlock(&crrInode, text, file, &tempSuperblock); err != nil {
		fmt.Println("Error: No se pudo escribir en el archivo:", err)
		return err
	}

	//reescribimos el superbloque
	if err := Utilities.WriteObject(file, tempSuperblock, int64(TempMBR.Partitions[index].Start)); err != nil {
		fmt.Println("Error: No se pudo escribir el Superblock:", err)
		return err
	}

	fmt.Println("======FIN MKGRP======")

	return nil
}

func activarGrupoDesactivado(file *os.File, lines []string, name string, tempSuperblock Structs.Superblock, crrInode Structs.Inode, cantidadG int) error {
	possition := 0

	// Iterar a través de las líneas para verificar las credenciales
	for _, line := range lines {
		fmt.Println("linea: ", line)
		if strings.Contains(line, "U") || !strings.Contains(line, name) && !strings.Contains(line, "G") {
			//sumamos la cantidad de caracteres de la linea y el salto de linea
			possition += len(line) + 1
			continue
		}

		//encontamos la linea que la contiene
		palabras := strings.Split(line, ",")
		found := false
		//vamos a buscar la posicion del numero para convertirlo en 0
		for _, palabra := range palabras {
			if !strings.Contains(palabra, name) {
				//sumamos la cantidad de caracteres de la palabra y la coma que le sigue
				possition += len(palabra) + 1
				continue
			}

			found = true
			//quitamos la coma anterior a la palabra
			possition -= 1
			//quitamos la "G" y la coma anterior a la G
			possition -= 2
			//posicionamos sobre el byte a cambiar
			possition -= 1
			//ahora en posicion tenemos justo el numero que queremos cambiar
			break
		}
		if found {
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
	//obtenemos la posicion relativa del bit a borrar
	//la posicion relativa es la posicion en el bloque
	//menos la cantidad de caracteres de las lineas anteriores
	relativePosition := int64(possition) % int64(tempSuperblock.S_block_size)

	crrFileBlock.B_content[relativePosition] = byte(cantidadG + 48)

	//reescribimos el bloque
	if err := Utilities.WriteObject(file, crrFileBlock, offset); err != nil {
		return err
	}

	return nil

}
