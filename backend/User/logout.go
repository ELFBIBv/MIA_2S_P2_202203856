package User

import (
	"fmt"
	"proyecto1/DiskManagement"
)

func Logout(LoggedPartitionID string) error {
	fmt.Println("======Start LOGOUT======")

	// Verificar si el usuario ya está logueado buscando en las particiones montadas
	mountedPartitions := DiskManagement.GetMountedPartitions()
	var partitionFound bool
	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.ID == LoggedPartitionID {
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		return fmt.Errorf("no se encontró ninguna partición montada con el ID '%s' proporcionado", LoggedPartitionID)
	}

	DiskManagement.MarkPartitionAsLoggedOut(LoggedPartitionID)

	fmt.Println("======End LOGOUT======")

	return nil
}
