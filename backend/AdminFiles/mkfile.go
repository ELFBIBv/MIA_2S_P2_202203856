package AdminFiles

import (
	"bufio"
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

func Mkfile(path string, r bool, size int, cont string, idPartition string) error {
	// Validar los parametros
	err, contenido := validatemkfile(path, size, cont, idPartition)
	if err != nil {
		return err
	}
	fmt.Println("======Start MKFILE======")
	fmt.Println("Path:", path)
	fmt.Println("Size:", size)
	fmt.Println("Contenido:", contenido)
	fmt.Println("ID Partition:", idPartition)

	err, mountedPartition, index := DiskManagement.GetMountedPartitionByID(idPartition)
	if err != nil {
		return err
	}

	// Abrir archivo binario
	file, err := Utilities.OpenFile(mountedPartition.Path)
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

	var tempSuperblock Structs.Superblock

	// Leer el Superblock desde el archivo binario
	if err := Utilities.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[index].Start)); err != nil {
		fmt.Println("Error: No se pudo leer el Superblock:", err)
		return fmt.Errorf("error al leer el Superblock: %w", err)
	}

	//Structs.PrintSuperblock(tempSuperblock)

	// Crear el archivo
	err = createFile(file, path, size, contenido, &tempSuperblock, index, r)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %w", err)
	}

	//Structs.PrintSuperblock(tempSuperblock)

	return nil
}

// Funcion para crear un archivo
func createFile(file *os.File, path string, size int, content string, sb *Structs.Superblock, posicionParticion int64, r bool) error {
	fmt.Println("\nCreando archivo:", file.Name()) // Imprimir el path del archivo

	parentDirs, destFile := Utilities.GetParentDirectories(path)
	fmt.Println("\nDirectorios padres:", parentDirs)
	fmt.Println("file destino:", destFile)

	if !strings.Contains(destFile, ".") {
		return errors.New("no se quiere crear un archivo")
	}

	// Obtener contenido por chunks
	chunks := Utilities.SplitStringIntoChunks(content)
	fmt.Println("\nChunks del contenido:", chunks)

	// Crear el archivo segun el path proporcionado
	if r {
		for quantity, parentDir := range parentDirs {
			err := CreateFolderInInode(sb, file, 0, parentDirs[:quantity], parentDir, 0, r)
			if err != nil {
				return err
			}
		}
	}
	err := createFileInInode(sb, file, 0, parentDirs, destFile, size, chunks, 0)
	if err != nil {
		return err
	}

	// Imprimir inodos y bloques
	err = FileSystem.PrintInodes(file, *sb)
	if err != nil {
		return fmt.Errorf("error al imprimir los inodos: %w", err)
	}
	err = FileSystem.PrintBlocksInOrder(file, *sb)
	if err != nil {
		return fmt.Errorf("error al imprimir los bloques: %w", err)
	}

	//obtenemos el mbr
	var mbr Structs.MRB
	//leemos el mbr
	err = Utilities.ReadObject(file, &mbr, 0)
	if err != nil {
		return fmt.Errorf("error al leer el mbr: %w", err)
	}

	// Serializar el superbloque
	//err = Utilities.WriteObject(file, sb, int64(mountedPartition.Part_start))
	err = Utilities.WriteObject(file, sb, int64(mbr.Partitions[posicionParticion].Start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}

// createFolderinode crea una carpeta en un inodo específico
func createFileInInode(sb *Structs.Superblock, file *os.File, inodeIndex int32, parentsDir []string, destFile string, fileSize int, fileContent []string, ParentInodeIndex int32) error {
	// Crear un nuevo inodo
	inode := &Structs.Inode{}
	// Deserializar el inodo
	err := Utilities.ReadObject(file, inode, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
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
		err := Utilities.ReadObject(file, &block, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
		if err != nil {
			return err
		}

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			// Obtener el contenido del bloque
			//content :=

			// Sí las carpetas padre no están vacías debereamos buscar la carpeta padre más cercana
			if len(parentsDir) != 0 {
				//fmt.Println("---------ESTOY  VISITANDO--------")

				// Si el contenido está vacío, salir
				if block.B_content[indexContent].B_inodo == -1 {
					break
				}

				// Obtenemos la carpeta padre más cercana
				parentDir := parentsDir[0]

				// Convertir B_name a string y eliminar los caracteres nulos
				contentName := strings.Trim(string(block.B_content[indexContent].B_name[:]), "\x00 ")
				// Convertir parentDir a string y eliminar los caracteres nulos
				parentDirName := strings.Trim(parentDir, "\x00 ")
				// Si el nombre del contenido coincide con el nombre de la carpeta padre
				if strings.EqualFold(contentName, parentDirName) {
					//fmt.Println("---------ESTOY  ENCONTRANDO--------")
					//el ParentsDir[1:] es para quitar el primer elemento de la lista
					err := createFileInInode(sb, file, block.B_content[indexContent].B_inodo, parentsDir[1:], destFile, fileSize, fileContent, inodeIndex)
					if err != nil {
						return err
					}
					return nil
				}
			} else {
				//fmt.Println("---------ESTOY  CREANDO--------")

				// Si el apuntador al inodo está ocupado, continuar con el siguiente
				if block.B_content[indexContent].B_inodo != -1 {
					// Convertir B_name a string y eliminar los caracteres nulos

					fmt.Println("Nombre del archivo: ", string(block.B_content[indexContent].B_name[:]))
					nameSinNulos := strings.Trim(string(block.B_content[indexContent].B_name[:]), "\x00 ")
					destinoClean := strings.TrimSpace(destFile)
					if strings.EqualFold(nameSinNulos, destinoClean) {
						return errors.New("el archivo ya existe")
					}
					continue
				}

				// Actualizar el contenido del bloque
				copy(block.B_content[indexContent].B_name[:], destFile)
				block.B_content[indexContent].B_inodo = sb.S_inodes_count

				// Serializar el bloque
				err = Utilities.WriteObject(file, block, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				//err = block.Serialize(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return fmt.Errorf("error al serializar el bloque %d: %w", blockIndex, err)
				}

				currentTime := time.Now()
				formattedDate := currentTime.Format("2006-01-02 15:04") //16 bytes

				// Crear el inodo del archivo
				fileInode := &Structs.Inode{
					I_uid:   1,
					I_gid:   1,
					I_size:  int32(fileSize),
					I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'1'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				copy(fileInode.I_atime[:], formattedDate)
				copy(fileInode.I_ctime[:], formattedDate)
				copy(fileInode.I_mtime[:], formattedDate)

				// Crear el bloques del archivo
				for i := 0; i < len(fileContent); i++ {
					// Actualizamos el inodo del archivo
					fileInode.I_block[i] = sb.S_blocks_count

					// Creamos el bloque del archivo
					fileBlock := Structs.Fileblock{
						B_content: [64]byte{},
					}
					// Copiamos el texto de usuarios en el bloque
					copy(fileBlock.B_content[:], fileContent[i])

					// Serializar el bloque del archivo
					err = Utilities.WriteObject(file, fileBlock, int64(sb.S_first_blo))
					//err = fileBlock.Serialize(file, int64(sb.S_first_blo))
					if err != nil {
						return fmt.Errorf("error al serializar el bloque del archivo: %w", err)
					}

					// Actualizar el bitmap de bloques
					//err = Utilities.WriteObject(file, byte('1'), int64(sb.S_bm_block_start+sb.S_first_blo))
					err = FileSystem.UpdateBitmapBlock(file, sb)
					//err = sb.UpdateBitmapBlock(file)
					if err != nil {
						return fmt.Errorf("error al actualizar el bitmap de bloques: %w", err)
					}

					// Actualizamos el superbloque
					sb.S_blocks_count++
					sb.S_free_blocks_count--
					sb.S_first_blo += sb.S_block_size
				}

				// Serializar el inodo de la carpeta+
				err = Utilities.WriteObject(file, *fileInode, int64(sb.S_inode_start+(sb.S_inodes_count*sb.S_inode_size)))
				//err = fileInode.Serialize(file, int64(sb.S_first_ino))
				if err != nil {
					return fmt.Errorf("error al serializar el inodo del archivo: %w", err)
				}

				// Actualizar el bitmap de inodos
				//err = sb.UpdateBitmapInode(file)
				err = Utilities.WriteObject(file, byte('1'), int64(sb.S_bm_inode_start+sb.S_inodes_count))
				if err != nil {
					return fmt.Errorf("error al actualizar el bitmap de inodos: %w", err)
				}

				// Actualizar el superbloque
				sb.S_inodes_count++
				sb.S_free_inodes_count--
				sb.S_first_ino += sb.S_inode_size

				return nil
			}
		}

	}
	return nil
}

func validatemkfile(path string, size int, cont string, idPartition string) (error, string) {
	if path == "" {
		return fmt.Errorf("path no especificado"), ""
	}

	contenido := ""
	if size < 0 {
		return fmt.Errorf("size no puede ser negativo"), ""
	}

	if cont == "" {
		for size > len(contenido) {
			contenido += "0123456789"
		}
		return nil, contenido[:size] // Recorta la cadena al tamaño exacto
	}

	if idPartition == "" {
		return errors.New("idPartition no especificado"), ""
	}

	// significa que tenemos que cont != "" por lo tanto esta tiene mas prioridad
	contenido = ""
	file, err := Utilities.OpenFile(cont)
	if err != nil {
		fmt.Println("Error: ", err)
		return fmt.Errorf("Error al abrir el archivo: %w", err), ""
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		contenido += scanner.Text() + "\n"
	}

	contenido = strings.Trim(contenido, "\n")

	return nil, contenido
}
