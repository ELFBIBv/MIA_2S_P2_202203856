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
	"strings"
	"time"
)

func Mkfs(id string, type_ string) error {
	fmt.Println("======INICIO MKFS======")
	fmt.Println("Id:", id)
	fmt.Println("Type:", type_)

	// Buscar la partición montada por ID
	var mountedPartition DiskManagement.MountedPartition
	var partitionFound bool

	for _, partitions := range DiskManagement.GetMountedPartitions() {
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
		return errors.New("partition no encontrada entre las particiones montadas")
	}

	//revisa si la particion esta montada
	if mountedPartition.Status != '1' { // Verifica si la partición está montada
		return errors.New("Particion'" + mountedPartition.Name + "' no montada")
	}

	// Abrir archivo binario
	//recordar que aqui vamos a abrir el disco
	file, err := Utilities.OpenFile(mountedPartition.Path)
	if err != nil {
		return err
	}

	var TempMBR Structs.MRB
	// Leer objeto desde archivo binario
	//vamos a leer el MBR
	if err = Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		return err
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
		return errors.New("Particion con ID'" + id + "' no encontrada entre las particiones del MBR")
	}
	Structs.PrintPartition(TempMBR.Partitions[index])

	//obtener el numero de inodos usando la formula
	//tamaño de la particion - tamaño del superbloque
	numerador := TempMBR.Partitions[index].Size - int32(binary.Size(Structs.Superblock{}))
	if numerador < 0 {
		return errors.New("tamaño de partición insuficiente para crear el sistema de archivos")
	}

	//factorizando el valor de n nos queda lo siguiente
	denominadorBase := 1 + 3 + int32(binary.Size(Structs.Inode{})) + 3*int32(binary.Size(Structs.Fileblock{}))
	n := numerador / denominadorBase

	fmt.Println("INODOS:", n)

	//obtener la fecha del sistema en formato YYYY-MM-DD HH:MM
	currentTime := time.Now()
	formattedDate := currentTime.Format("2006-01-02 15:04") //16 bytes

	// Crear el Superblock con todos los campos calculados
	var newSuperblock Structs.Superblock
	newSuperblock.S_filesystem_type = 2 // EXT2
	newSuperblock.S_inodes_count = 0
	newSuperblock.S_blocks_count = 0
	newSuperblock.S_free_blocks_count = 3 * n
	newSuperblock.S_free_inodes_count = n
	copy(newSuperblock.S_mtime[:], formattedDate)
	copy(newSuperblock.S_umtime[:], formattedDate)
	newSuperblock.S_mnt_count = 1
	newSuperblock.S_magic = 0xEF53
	newSuperblock.S_inode_size = int32(binary.Size(Structs.Inode{}))
	newSuperblock.S_block_size = int32(binary.Size(Structs.Fileblock{}))

	// Calcula las posiciones de inicio
	newSuperblock.S_bm_inode_start = TempMBR.Partitions[index].Start + int32(binary.Size(Structs.Superblock{}))
	newSuperblock.S_bm_block_start = newSuperblock.S_bm_inode_start + n
	newSuperblock.S_inode_start = newSuperblock.S_bm_block_start + 3*n
	newSuperblock.S_block_start = newSuperblock.S_inode_start + n*newSuperblock.S_inode_size

	//comienzo de los inodos y bloques libres
	newSuperblock.S_first_ino = newSuperblock.S_inode_start
	newSuperblock.S_first_blo = newSuperblock.S_block_start

	//creamos el sistema ext2
	//func Create_ext2(n int32, partition Structs.Partition, newSuperblock Structs.Superblock, date string, file *os.File) error {

	err = createExt2(n, TempMBR.Partitions[index], newSuperblock, file)
	if err != nil {
		return err
	}

	// Cerrar archivo binario
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error: No se pudo cerrar el archivo:", err)
		}
	}(file)

	fmt.Println("======FIN MKFS======")
	return nil
}

func createExt2(n int32, partition Structs.Partition, newSuperblock Structs.Superblock, file *os.File) error {
	fmt.Println("======Start CREATE EXT2======")
	fmt.Println("INODOS:", n)

	// Imprimir Superblock inicial
	//Structs.PrintSuperblock(newSuperblock)

	// Escribe los bitmaps de inodos y bloques en el archivo
	for i := '0'; i < n; i++ {
		if err := Utilities.WriteObject(file, byte('0'), int64(newSuperblock.S_bm_inode_start+i)); err != nil {
			fmt.Println("Error: ", err)
			return err
		}
	}

	for i := int32(0); i < 3*n; i++ {
		if err := Utilities.WriteObject(file, byte('0'), int64(newSuperblock.S_bm_block_start+i)); err != nil {
			fmt.Println("Error: ", err)
			return err
		}
	}

	// Inicializa inodos y bloques con valores predeterminados
	if err := FileSystem.InitInodesAndBlocks(n, newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	// Crea la carpeta raíz y el archivo users.txt
	if err := createRootAndUsersFile(&newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	// Marca los primeros inodos y bloques como usados
	if err := markUsedInodesAndBlocks(&newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	fmt.Print("el superbloque se escribio en el byte:" + string(partition.Start))

	// Escribe el superbloque actualizado al archivo
	if err := Utilities.WriteObject(file, newSuperblock, int64(partition.Start)); err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	// Leer e imprimir los inodos después de formatear
	fmt.Println("====== Imprimiendo Inodos ======")
	for i := int32(0); i < n; i++ {
		var inode Structs.Inode
		offset := int64(newSuperblock.S_inode_start + i*int32(binary.Size(Structs.Inode{})))
		if err := Utilities.ReadObject(file, &inode, offset); err != nil {
			fmt.Println("Error al leer inodo: ", err)
			return err
		}
		if i < 5 {
			Structs.PrintInode(inode)
		}
	}

	// Leer e imprimir los Folderblocks y Fileblocks después de formatear
	fmt.Println("====== Imprimiendo Folderblocks y Fileblocks ======")

	// Imprimir Folderblocks
	for i := int32(0); i < 1; i++ {
		var folderblock Structs.Folderblock
		offset := int64(newSuperblock.S_block_start + i*int32(binary.Size(Structs.Folderblock{})))
		if err := Utilities.ReadObject(file, &folderblock, offset); err != nil {
			fmt.Println("Error al leer Folderblock: ", err)
			return err
		}
		Structs.PrintFolderblock(folderblock)
	}

	// Imprimir Fileblocks
	for i := int32(0); i < 1; i++ {
		var fileblock Structs.Fileblock
		offset := int64(newSuperblock.S_block_start + int32(binary.Size(Structs.Folderblock{})) + i*int32(binary.Size(Structs.Fileblock{})))
		if err := Utilities.ReadObject(file, &fileblock, offset); err != nil {
			fmt.Println("Error al leer Fileblock: ", err)
			return err
		}
		Structs.PrintFileblock(fileblock)
	}

	// Imprimir el Superblock final
	Structs.PrintSuperblock(newSuperblock)

	fmt.Println("======End CREATE EXT2======")

	return nil
}

// Función auxiliar para crear la carpeta raíz y el archivo users.txt
func createRootAndUsersFile(newSuperblock *Structs.Superblock, file *os.File) error {
	//var Inode0, Inode1 Structs.Inode

	//obtenemos la fecha actual
	currentTime := time.Now()
	formattedDate := currentTime.Format("2006-01-02 15:04") //16 bytes

	// Copiar la fecha formateada al inodo
	var currentTimeBytes [16]byte
	copy(currentTimeBytes[:], formattedDate)

	// Crear el inodo de la carpeta
	Inode0 := Structs.Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  0,
		I_atime: currentTimeBytes,
		I_ctime: currentTimeBytes,
		I_mtime: currentTimeBytes,
		I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'0'},
		I_perm:  [3]byte{'6', '6', '4'},
	}

	// Crear el inodo del archivo
	Inode1 := Structs.Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  0,
		I_atime: currentTimeBytes,
		I_ctime: currentTimeBytes,
		I_mtime: currentTimeBytes,
		I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'1'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	//FileSystem.InitInode(&Inode0, date)
	//FileSystem.InitInode(&Inode1, date)

	Inode0.I_block[0] = 0
	Inode1.I_block[0] = 1

	// Asignar el tamaño real del contenido
	data := "1,G,root\n1,U,root,root,123"
	actualSize := int32(len(data))
	Inode1.I_size = actualSize // Esto ahora refleja el tamaño real del contenido

	Fileblock1 := Structs.Fileblock{
		B_content: [64]byte{},
	}

	copy(Fileblock1.B_content[:], data) // Copia segura de datos a Fileblock

	// Crear el Folderblock de la carpeta raíz
	var temp [12]byte
	copy(temp[:], "users.txt")

	Folderblock0 := Structs.Folderblock{
		B_content: [4]Structs.Content{
			{B_name: [12]byte{'.'}, B_inodo: 0},
			{B_name: [12]byte{'.', '.'}, B_inodo: 0},
			{B_name: temp, B_inodo: 1},
			{B_name: [12]byte{'-', '1'}, B_inodo: -1},
		},
	}
	//
	//var Folderblock0 Structs.Folderblock
	//Folderblock0.B_content[0].B_inodo = 0
	//copy(Folderblock0.B_content[0].B_name[:], ".")
	//Folderblock0.B_content[1].B_inodo = 0
	//copy(Folderblock0.B_content[1].B_name[:], "..")
	//Folderblock0.B_content[2].B_inodo = 1
	//copy(Folderblock0.B_content[2].B_name[:], "users.txt")

	//cambiamos el inicio del bloque e inodo libre
	newSuperblock.S_first_ino += 2 * int32(binary.Size(Structs.Inode{}))
	newSuperblock.S_first_blo += 2 * int32(binary.Size(Structs.Fileblock{}))
	newSuperblock.S_free_blocks_count -= 2
	newSuperblock.S_free_inodes_count -= 2
	newSuperblock.S_inodes_count += 2
	newSuperblock.S_blocks_count += 2

	// Escribir los inodos y bloques en las posiciones correctas
	if err := Utilities.WriteObject(file, Inode0, int64(newSuperblock.S_inode_start)); err != nil {
		return err
	}
	if err := Utilities.WriteObject(file, Inode1, int64(newSuperblock.S_inode_start+int32(binary.Size(Structs.Inode{})))); err != nil {
		return err
	}
	if err := Utilities.WriteObject(file, Folderblock0, int64(newSuperblock.S_block_start)); err != nil {
		return err
	}
	if err := Utilities.WriteObject(file, Fileblock1, int64(newSuperblock.S_block_start+int32(binary.Size(Structs.Folderblock{})))); err != nil {
		return err
	}

	return nil
}

// Función auxiliar para marcar los inodos y bloques usados
func markUsedInodesAndBlocks(newSuperblock *Structs.Superblock, file *os.File) error {
	if err := Utilities.WriteObject(file, byte('1'), int64(newSuperblock.S_bm_inode_start)); err != nil {
		return err
	}
	if err := Utilities.WriteObject(file, byte('1'), int64(newSuperblock.S_bm_inode_start+1)); err != nil {
		return err
	}
	if err := Utilities.WriteObject(file, byte('1'), int64(newSuperblock.S_bm_block_start)); err != nil {
		return err
	}
	if err := Utilities.WriteObject(file, byte('1'), int64(newSuperblock.S_bm_block_start+1)); err != nil {
		return err
	}
	return nil
}
