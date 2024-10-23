package DiskManagement

import (
	"errors"
	"fmt"
	"proyecto1/Structs"
	"proyecto1/Utilities"
	"strings"
)

// MountedPartition Estructura para representar una partición montada
type MountedPartition struct {
	Path     string
	Name     string
	ID       string
	Status   byte // 0: no montada, 1: montada
	LoggedIn bool // true: usuario ha iniciado sesión, false: no ha iniciado sesión
}

// Mapa para almacenar las particiones montadas, organizadas por disco
var MountedPartitions = make(map[string][]MountedPartition)

// PrintMountedPartitions Función para imprimir las particiones montadas
func PrintMountedPartitions() {
	fmt.Println("Particiones montadas:")

	if len(MountedPartitions) == 0 {
		fmt.Println("No hay particiones montadas.")
		return
	}

	for diskID, partitions := range MountedPartitions {
		fmt.Printf("Disco ID: %s\n", diskID)
		for _, partition := range partitions {
			loginStatus := "No"
			if partition.LoggedIn {
				loginStatus = "Sí"
			}
			fmt.Printf(" - Partición Name: %s, ID: %s, Path: %s, Status: %c, LoggedIn: %s\n",
				partition.Name, partition.ID, partition.Path, partition.Status, loginStatus)
		}
	}
	fmt.Println("")
}

// GetMountedPartitions Función para obtener las particiones montadas
func GetMountedPartitions() map[string][]MountedPartition {
	return MountedPartitions
}

// MarkPartitionAsLoggedIn Función para marcar una partición como logueada
func MarkPartitionAsLoggedIn(id string) {
	for diskID, partitions := range MountedPartitions {
		for i, partition := range partitions {
			if partition.ID == id {
				MountedPartitions[diskID][i].LoggedIn = true
				fmt.Printf("Partición con ID %s marcada como logueada.\n", id)
				return
			}
		}
	}
	fmt.Printf("No se encontró la partición con ID %s para marcarla como logueada.\n", id)
}

func MarkPartitionAsLoggedOut(id string) {
	for diskID, partitions := range MountedPartitions {
		for i, partition := range partitions {
			if partition.ID == id {
				MountedPartitions[diskID][i].LoggedIn = false
				fmt.Printf("Partición con ID %s marcada como deslogueada.\n", id)
				return
			}
		}
	}
	fmt.Printf("No se encontró la partición con ID %s para marcarla como deslogueada.\n", id)
}

func GetMountedPartitionByID(id string) (error, MountedPartition, int64) {
	var partitionFound bool
	var mountedPartition MountedPartition

	for _, partitions := range GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.ID == id {
				mountedPartition = partition
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	//revisa se la particion fue encontrada
	if !partitionFound {
		return errors.New("partition no encontrada entre las particiones montadas"), mountedPartition, -1
	}

	//revisa si la particion esta montada
	if mountedPartition.Status != '1' { // Verifica si la partición está montada
		return errors.New("Particion'" + mountedPartition.Name + "' no montada"), mountedPartition, -1
	}

	// Abrir archivo binario
	file, err := Utilities.OpenFile(mountedPartition.Path)
	if err != nil {
		return err, mountedPartition, -1
	}

	var TempMBR Structs.MRB
	// Leer objeto desde archivo binario
	//vamos a leer el MBR
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		return err, mountedPartition, -1
	}

	// Imprimir objeto
	Structs.PrintMBR(TempMBR)

	fmt.Println("-------------")

	var index = -1
	// Iterar sobre las particiones para encontrar la que tiene el nombre correspondiente
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), id) {
				//index indica la posicion de la particion
				index = i
				break
			}
		}
	}

	if index == -1 {
		return errors.New("Particion con ID'" + id + "' no encontrada entre las particiones del MBR"), mountedPartition, -1
	}

	return nil, mountedPartition, int64(index)
}
