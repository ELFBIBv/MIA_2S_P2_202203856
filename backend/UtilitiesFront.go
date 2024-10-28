package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"proyecto1/FileSystem"
	"proyecto1/Structs"
	"proyecto1/Utilities"
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
	file, err := Utilities.OpenFile(path)
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
	err = Utilities.ReadObject(file, &mbr, 0) // Leer desde la posición 0
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

// la estructura de los archivos de salida
type files struct {
	Name string
	Type string
}

func recolectFiles(filePath string, diskPath string, partitionName string) (error, []files) {
	// Abrir el archivo binario

	file, err := Utilities.OpenFile(diskPath) //abrimos el disco
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo en la ruta:", diskPath)
		return err, nil
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
		return err, nil
	}

	var found = false
	// Convertir el nombre a comparar a un arreglo de bytes de longitud fija
	nameBytes := [16]byte{}
	copy(nameBytes[:], partitionName)
	Structs.PrintMBR(TempMBR)
	index := 0
	for i := 0; i < 4; i++ {
		// Verificar si el nombre coincide
		if bytes.Equal(TempMBR.Partitions[i].Name[:], nameBytes[:]) {
			found = true
			index = i
			fmt.Println("posicion de particion: ", i)
			break
		}
	}

	if !found {
		return errors.New("No se encontró la partición con nombre: " + partitionName), nil
	}
	// ahora que ya tenemos la particion vamos a buscar los archivos
	//primero vamos a conseguir el mbr de la particion
	var MBRTemporal Structs.MRB
	// Leer el MBR desde el archivo
	err = Utilities.ReadObject(file, &MBRTemporal, 0) // Leer desde la posición 0
	if err != nil {
		return err, nil
	}

	Structs.PrintMBR(MBRTemporal)
	fmt.Printf("posicion sb: %d\n", MBRTemporal.Partitions[index].Start)

	//vamos a obtener el superbloque
	var sb = Structs.Superblock{}
	if err := Utilities.ReadObject(file, &sb, int64(MBRTemporal.Partitions[index].Start)); err != nil {
		fmt.Println("Error REP SB: Error al leer el SuperBloque.")
		return err, nil
	}

	Structs.PrintSuperblock(sb)

	// este es el numero de inodo del archivo
	inode_number := FileSystem.SearchInodeByPath(filePath, file, sb)
	if inode_number == -1 {
		fmt.Print("Error REP: No se encontró el archivo")
		return errors.New("Error REP: No se encontró el archivo"), nil
	}

	//ahora vamos a leer el inodo
	var inode = Structs.Inode{}
	posicion := sb.S_inode_start + inode_number*sb.S_inode_size
	if err := Utilities.ReadObject(file, &inode, int64(posicion)); err != nil {
		fmt.Printf("Error REP: Error al leer el Inodo %d\n", inode_number)
		fmt.Printf("posicion del inodo: %d\n", posicion)
		fmt.Printf("inicio de inodos: %d\n", sb.S_inode_start)
		fmt.Printf("tamaño del inodo: %d\n", sb.S_inode_size)
		fmt.Printf("tamaño del inodo 2: %d\n", inode.I_size)
		return err, nil
	}

	//primero vamos a ver que tipo de inodo es
	if inode.I_type[0] == '1' {
		//es un archivo
		fmt.Println("es un archivo")
		return nil, []files{{Name: "error", Type: "file"}}
	}
	//--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
	//vamos a leer todos los bloques del inodo
	fmt.Print("Inodo: ")
	Structs.PrintInode(inode)

	arrayFiles := []files{}
	for _, blockIndex := range inode.I_block {
		// Si el bloque no existe, salir
		if blockIndex == -1 {
			fmt.Println("Error: No se encontró el bloque")
			break
		}
		// Crear un nuevo bloque de carpeta
		block := Structs.Folderblock{}

		// Deserializar el bloque
		//err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
		err := Utilities.ReadObject(file, &block, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			fmt.Println("Error al leer el bloque:", err)
			return err, nil
		}

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			// Obtener el contenido del bloque
			var file files
			nameTemp := block.B_content[indexContent].B_name
			name := string(bytes.Trim(nameTemp[:], "\x00"))

			// Elimina espacios internos y tabulaciones
			name = strings.ReplaceAll(name, " ", "")
			name = strings.ReplaceAll(name, "\t", "")
			fmt.Println("Nombre del archivo: ", name)

			if block.B_content[indexContent].B_inodo == -1 {
				break
			}
			file.Name = name

			if strings.Contains(name, ".") {
				// Es un archivo
				file.Type = "file"
			} else {
				// Es una carpeta
				file.Type = "folder"
			}
			arrayFiles = append(arrayFiles, file)
		}
	}
	return nil, arrayFiles
}
