package User

import (
	"encoding/binary"
	"errors"
	"fmt"
	"proyecto1/DiskManagement"
	"proyecto1/FileSystem"
	"proyecto1/Structs"
	"proyecto1/Utilities"
	"strconv"
	"strings"
)

func Mkusr(user string, pass string, grp string, id string) error {
	fmt.Println("======INICIO MKUSR======")
	fmt.Println("User:", user)
	fmt.Println("Pass:", pass)
	fmt.Println("Grp:", grp)

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
					return errors.New(
						"la partición no está montada")
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
	groupFound := false

	//vamos a comprobar de que no el grupo al que se le quiere asignar el usuario no este desactivado
	for _, line := range lines {
		palabras := strings.Split(line, ",")
		if palabras[0] == "0" && palabras[1] == "G" && palabras[2] == grp {
			fmt.Println("linea:", line)
			fmt.Println("Error: El grupo al que se le quiere asignar el usuario está desactivado")
			return errors.New("el grupo al que se le quiere asignar el usuario está desactivado")
		}

		//revisamos que el grupo si exista
		if palabras[2] == grp && palabras[1] == "G" {
			groupFound = true
		}
	}

	if !groupFound {
		fmt.Println("Error: El grupo al que se le quiere asignar el usuario no existe")
		return errors.New("el grupo al que se le quiere asignar el usuario no existe")
	}

	//revisamos que el usuario no se repita
	for _, line := range lines {
		palabras := strings.Split(line, ",")
		if !strings.Contains(palabras[0], "0") && strings.Contains(palabras[1], "U") {
			if strings.Contains(palabras[3], user) {
				fmt.Println("Error: El usuario ya existe")
				return errors.New("el usuario ya existe")
			}
		}
	}

	cantidadU := 1

	// Iterar a través de las líneas para verificar las credenciales
	for {
		found := false
		for _, line := range lines {
			palabras := strings.Split(line, ",")
			if palabras[0] == strconv.Itoa(cantidadU) && palabras[1] == "U" {
				found = true
				break
			}
		}
		if !found {
			break
		}
		cantidadU++
		if cantidadU > 9 {
			break
		}
	}

	if cantidadU > 9 {
		return errors.New("no se pueden tener más de 9 grupos")
	}

	// Crear nuevo grupo
	text := fmt.Sprintf("%d,U,%s,%s,%s", cantidadU, grp, user, pass)

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

	fmt.Println("======FIN MKUSR======")

	return nil
}
