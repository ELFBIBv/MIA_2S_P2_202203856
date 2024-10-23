package AdminFiles

import (
	"errors"
	"fmt"
	"log"
	"os"
	"proyecto1/DiskManagement"
	"proyecto1/FileSystem"
	"proyecto1/Structs"
	"proyecto1/Utilities"
	"strings"
	"time"
)

func Mkdir(path string, p bool, idPartition string) error {

	//validar parametros
	err := validateMkdir(path, idPartition)
	if err != nil {
		return fmt.Errorf("error al validar los parametros: %w", err)
	}
	fmt.Println("======Start MKDIR======")
	fmt.Println("path:", path)
	fmt.Println("p:", p)
	fmt.Println("idPartition:", idPartition)

	var mountedPartition DiskManagement.MountedPartition
	var index int64 = 0

	err, mountedPartition, index = DiskManagement.GetMountedPartitionByID(idPartition)
	if err != nil {
		return err
	}

	//abrimos el archivo
	file, err := Utilities.OpenFile(mountedPartition.Path)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %w", err)
	}

	var TempMBR Structs.MRB
	// Leer el MBR desde el archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return err
	}

	var partitionSuperblock Structs.Superblock

	partition := TempMBR.Partitions[index]
	//
	if err := Utilities.ReadObject(file, &partitionSuperblock, int64(partition.Start)); err != nil {
		fmt.Println("Error: No se pudo leer el Superblock:", err)
		return fmt.Errorf("error al leer el Superblock: %w", err)
	}

	//Structs.PrintSuperblock(partitionSuperblock)

	// Crear el directorio
	err = createDirectory(path, &partitionSuperblock, file, p)
	if err != nil {
		return fmt.Errorf("error al crear el directorio: %w", err)
	}

	if err = Utilities.WriteObject(file, &partitionSuperblock, int64(partition.Start)); err != nil {
		return fmt.Errorf("error al escribir el superbloque: %w", err)
	}

	fmt.Println("Se creó el directorio correctamente")

	defer func() {
		fmt.Println("Cerrando el archivo")
		if file != nil {
			err := file.Close()
			if err != nil {
				log.Println("Error al cerrar el archivo:", err)
			}
		}
	}()

	return err
}

func createDirectory(dirPath string, sb *Structs.Superblock, file *os.File, p bool) error {
	fmt.Println("\nCreando directorio:", dirPath)

	parentDirs, destDir := Utilities.GetParentDirectories(dirPath)
	fmt.Println("\nDirectorios padres:", parentDirs)
	fmt.Println("Directorio destino:", destDir)

	// Crear el directorio segun el path proporcionado
	if p {
		for quantity, parentDir := range parentDirs {
			err := CreateFolderInInode(sb, file, 0, parentDirs[:quantity], parentDir, 0, p)
			if err != nil {
				return err
			}
		}
	}
	err := CreateFolderInInode(sb, file, 0, parentDirs, destDir, 0, p)
	if err != nil {
		return err
	}

	// Imprimir inodos y bloques
	//fmt.Println("\nImprimiendo inodos y bloques")
	//err = FileSystem.PrintInodes(file, *sb)
	//if err != nil {
	//	fmt.Println("Error al imprimir inodos:", err)
	//	return err
	//}
	//err = FileSystem.PrintBlocksInOrder(file, *sb)
	//if err != nil {
	//	fmt.Println("Error al imprimir bloques:", err)
	//	return err
	//}

	return nil
}

// createFolderInInode crea una carpeta en un inodo específico
func CreateFolderInInode(sb *Structs.Superblock, file *os.File, inodeIndex int32, parentsDir []string, destDir string, ParentInodeIndex int32, r bool) error {
	fmt.Println("\nCreando carpeta en el inodo:", inodeIndex)
	// Crear un nuevo inodo
	inode := &Structs.Inode{}
	// Deserializar el inodo
	err := Utilities.ReadObject(file, inode, int64(sb.S_inode_start+inodeIndex*sb.S_inode_size))
	if err != nil {
		fmt.Println("Error al leer el inodo:", err)
		return err
	}
	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] != '0' {
		fmt.Println("El inodo no es de tipo carpeta, ", inode.I_type[0])
		Structs.PrintInode(*inode)
		return errors.New("el inodo no es de tipo carpeta")
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for possition, blockIndex := range inode.I_block {
		// Si el bloque no existe, salir
		if blockIndex == -1 {
			if len(parentsDir) != 0 {
				//aqui no crearemos bloques padre
				fmt.Println("El bloque no existe y no se puede crear por ser de ruta padre")
				return errors.New("el bloque no existe y no se puede crear por ser de ruta padre")
			} else {
				fmt.Println("Creando bloque de carpeta")
				// Crear el folderblock
				block := Structs.Folderblock{
					B_content: [4]Structs.Content{
						{B_name: [12]byte{'.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'.', '.'}, B_inodo: ParentInodeIndex},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}

				// Serializar el folderblock
				err = Utilities.WriteObject(file, &block, int64(sb.S_first_blo))
				if err != nil {
					fmt.Println("Error al escribir el bloque de la carpeta:", err)
					return err
				}

				// Actualizar el bitmap de bloques
				err = FileSystem.UpdateBitmapBlock(file, sb)
				if err != nil {
					fmt.Println("Error al actualizar el bitmap de bloques:", err)
					return err
				}

				//apuntamos el blockIndex al bloque creado
				blockIndex = sb.S_blocks_count
				// Actualizar el inodo
				inode.I_block[possition] = blockIndex

				// Actualizar el superbloque
				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size

				// Serializar el inodo
				err = Utilities.WriteObject(file, inode, int64(sb.S_inode_start+inodeIndex*sb.S_inode_size))
				if err != nil {
					fmt.Println("Error al escribir el inodo:", err)
					return err
				}
				fmt.Printf("Bloque de carpeta '%d'\n", blockIndex)
			}

		}
		// Crear un nuevo bloque de carpeta
		block := Structs.Folderblock{}

		// Deserializar el bloque
		//err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
		err := Utilities.ReadObject(file, &block, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			fmt.Println("Error al leer el bloque:", err)
			return err
		}

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			// Obtener el contenido del bloque
			fmt.Printf("Content %d: Name: %s, Inodo: %d\n", indexContent, string(block.B_content[indexContent].B_name[:]), block.B_content[indexContent].B_inodo)

			// Sí las carpetas padre no están vacías debereamos buscar la carpeta padre más cercana
			if len(parentsDir) != 0 {
				fmt.Println("---------ESTOY  VISITANDO--------")

				// Si el contenido está vacío, salir
				if block.B_content[indexContent].B_inodo == -1 {
					return errors.New("la carpeta padre'" + parentsDir[0] + "' no existe")
				}

				// Obtenemos la carpeta padre más cercana
				parentDir := parentsDir[0]

				// Convertir B_name a string y eliminar los caracteres nulos
				contentName := strings.Trim(string(block.B_content[indexContent].B_name[:]), "\x00 ")
				// Convertir parentDir a string y eliminar los caracteres nulos
				parentDirName := strings.Trim(parentDir, "\x00 ")
				// Si el nombre del contenido coincide con el nombre de la carpeta padre
				if strings.EqualFold(contentName, parentDirName) {
					//fmt.Println("---------LA ENCONTRÉ-------")
					// Si son las mismas, entonces entramos al inodo que apunta el bloque
					fmt.Println("el inodo apuntado por el bloque es:", block.B_content[indexContent].B_inodo)
					err := CreateFolderInInode(sb, file, block.B_content[indexContent].B_inodo, parentsDir[1:], destDir, inodeIndex, r)
					if err != nil {
						fmt.Println("Error al crear la carpeta en el inodo:", block.B_content[indexContent].B_inodo)
						return err
					}
					return nil
				}
			} else {
				fmt.Println("---------ESTOY  CREANDO--------")

				// Si el apuntador al inodo está ocupado, continuar con el siguiente
				fmt.Println("estoy creando el directorio en el byte:", int64(sb.S_first_ino))
				if block.B_content[indexContent].B_inodo != -1 {
					//verificar si la carpeta ya existe
					fmt.Println("el inodo apuntado por el bloque es:", block.B_content[indexContent].B_inodo, "pasa al siguiente")
					name := strings.Trim(string(block.B_content[indexContent].B_name[:]), "\x00 ")
					if name == destDir {
						if r {
							return nil
						}
						return errors.New("la carpeta '" + destDir + "' ya existe")
					}
					continue
				}

				// Actualizar el contenido del bloque
				copy(block.B_content[indexContent].B_name[:], destDir)
				block.B_content[indexContent].B_inodo = sb.S_inodes_count

				// Actualizar el bloque
				//block.B_content[indexContent]. = content

				// Serializar el bloque
				fmt.Println("la posicion del bloque es:", blockIndex)
				fmt.Println("el bloque se quiere escribir en el byte:", int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				err = Utilities.WriteObject(file, block, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					fmt.Println("Error al escribir el bloque:", err)
					return err
				}

				//obtenemos la fecha actual
				currentTime := time.Now()
				formattedDate := currentTime.Format("2006-01-02 15:04") //16 bytes

				// Crear el inodo del archivo
				fileInode := &Structs.Inode{
					I_uid:   1,
					I_gid:   1,
					I_size:  0,
					I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'0'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				copy(fileInode.I_atime[:], formattedDate)
				copy(fileInode.I_ctime[:], formattedDate)
				copy(fileInode.I_mtime[:], formattedDate)

				// Serializar el inodo de la carpeta
				err = Utilities.WriteObject(file, *fileInode, int64(sb.S_first_ino))
				if err != nil {
					fmt.Println("Error al escribir el inodo:", err)
					return err
				}

				// Actualizar el bitmap de inodos
				err = FileSystem.UpdateBitmapInode(file, sb)
				if err != nil {
					fmt.Println("Error al actualizar el bitmap de inodos:", err)
					return err
				}

				// Actualizar el superbloque
				sb.S_inodes_count++
				sb.S_free_inodes_count--
				sb.S_first_ino += sb.S_inode_size

				// Crear el bloque de la carpeta
				//folderBlock := &Structs.Folderblock{
				//	B_content: [4]Structs.Content{
				//		{B_name: [12]byte{'.'}, B_inodo: block.B_content[indexContent].B_inodo},
				//		{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
				//		{B_name: [12]byte{'-'}, B_inodo: -1},
				//		{B_name: [12]byte{'-'}, B_inodo: -1},
				//	},
				//}

				// Serializar el bloque de la carpeta
				//err = folderBlock.Serialize(path, int64(sb.S_first_blo))
				//fmt.Println("el bloque se quiere escribir en el byte (2):", int64(sb.S_first_blo))
				//err = Utilities.WriteObject(file, *folderBlock, int64(sb.S_first_blo))
				//if err != nil {
				//	fmt.Println("Error al escribir el bloque de la carpeta:", err)
				//	return err
				//}

				// Actualizar el bitmap de bloques
				//err = sb.UpdateBitmapBlock(path)
				err = FileSystem.UpdateBitmapBlock(file, sb)
				if err != nil {
					fmt.Println("Error al actualizar el bitmap de bloques:", err)
					return err
				}

				// Actualizar el superbloque
				//sb.S_blocks_count++
				//sb.S_free_blocks_count--
				//sb.S_first_blo += sb.S_block_size

				return nil
			}
		}

	}
	return nil
}

func validateMkdir(path string, idPartition string) error {
	if path == "" {
		return fmt.Errorf("path es obligatorio")
	}

	if idPartition == "" {
		return fmt.Errorf("idPartition es obligatorio")
	}

	return nil
}
