package FileSystem

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"proyecto1/Structs"
	"proyecto1/Utilities"
	"strings"
)

// InitInodesAndBlocks Función auxiliar para inicializar inodos y bloques
func InitInodesAndBlocks(n int32, newSuperblock Structs.Superblock, file *os.File) error {
	var newInode Structs.Inode
	for i := int32(0); i < 15; i++ {
		newInode.I_block[i] = -1
	}

	for i := int32(0); i < n; i++ {
		if err := Utilities.WriteObject(file, newInode, int64(newSuperblock.S_inode_start+i*int32(binary.Size(Structs.Inode{})))); err != nil {
			return err
		}
	}

	var newFileblock Structs.Fileblock
	for i := int32(0); i < 3*n; i++ {
		if err := Utilities.WriteObject(file, newFileblock, int64(newSuperblock.S_block_start+i*int32(binary.Size(Structs.Fileblock{})))); err != nil {
			return err
		}
	}

	return nil
}

// InitInode Función auxiliar para inicializar un inodo
func InitInode(inode *Structs.Inode, date string) {
	inode.I_uid = 1
	inode.I_gid = 1
	inode.I_size = 0
	copy(inode.I_atime[:], date)
	copy(inode.I_ctime[:], date)
	copy(inode.I_mtime[:], date)
	copy(inode.I_perm[:], "664")

	for i := int32(0); i < 15; i++ {
		inode.I_block[i] = -1
	}
}

func SearchInodeByPath(path string, file *os.File, tempSuperblock Structs.Superblock) int32 {
	fmt.Println("======Start BUSQUEDA INICIAL ======")
	fmt.Println("path:", path)

	// split the path by /
	TempStepsPath := strings.Split(path, "/")
	StepsPath := TempStepsPath[1:]

	fmt.Println("StepsPath:", StepsPath, "len(StepsPath):", len(StepsPath))
	for _, step := range StepsPath {
		fmt.Println("step:", step)
	}

	var Inode0 Structs.Inode
	// Read object from bin file
	if err := Utilities.ReadObject(file, &Inode0, int64(tempSuperblock.S_inode_start)); err != nil {
		return -1
	}

	fmt.Println("======End BUSQUEDA INICIAL======")

	return searchInodeByPath_rec(StepsPath, Inode0, file, tempSuperblock)
}

func searchInodeByPath_rec(StepsPath []string, Inode Structs.Inode, file *os.File, tempSuperblock Structs.Superblock) int32 {
	fmt.Println("======Start BUSQUEDA INODO POR PATH======")
	index := int32(0)
	//SearchedName := strings.Replace(pop(&StepsPath), " ", "", -1)
	SearchedName := strings.Replace(StepsPath[0], " ", "", -1) // get the first element of the array
	StepsPath = StepsPath[1:]                                  // remove the first element of the array

	fmt.Println("========== SearchedName:", SearchedName)

	// Iterate over i_blocks from Inode
	for _, block := range Inode.I_block {
		if block != -1 {
			if index < 13 {
				//CASO DIRECTO

				var crrFolderBlock Structs.Folderblock
				// Read object from bin file
				if err := Utilities.ReadObject(file, &crrFolderBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(Structs.Folderblock{})))); err != nil {
					return -1
				}

				for _, folder := range crrFolderBlock.B_content {
					// fmt.Println("Folder found======")
					fmt.Println("Folder === Name:", string(folder.B_name[:]), "B_inodo", folder.B_inodo)

					if strings.Contains(string(folder.B_name[:]), SearchedName) {

						//fmt.Println("len(StepsPath)", len(StepsPath), "StepsPath", StepsPath)
						if len(StepsPath) == 0 {
							fmt.Println("Folder found======")
							return folder.B_inodo
						} else {
							fmt.Println("NextInode======")
							var NextInode Structs.Inode
							// Read object from bin file
							if err := Utilities.ReadObject(file, &NextInode, int64(tempSuperblock.S_inode_start+folder.B_inodo*int32(binary.Size(Structs.Inode{})))); err != nil {
								return -1
							}
							return searchInodeByPath_rec(StepsPath, NextInode, file, tempSuperblock)
						}
					}
				}

			} else {
				fmt.Print("indirectos")
			}
		}
		index++
	}

	fmt.Println("======End BUSQUEDA INODO POR PATH======")
	return 0
}

func GetInodeFileData(Inode Structs.Inode, file *os.File, tempSuperblock Structs.Superblock) (string, error) {
	fmt.Println("======Start CONTENIDO DEL BLOQUE======")
	index := int32(0)
	// define content as a string
	var content string

	if Inode.I_type[0] != '1' {
		fmt.Println("Inode no es un archivo, es de tipo:", Inode.I_type[0])
		return "", errors.New("inode no es un archivo")
	}

	// Iterate over i_blocks from Inode
	for _, block := range Inode.I_block {
		if block != -1 {
			//Dentro de los directos
			if index < 13 {
				var crrFileBlock Structs.Fileblock
				// Read object from bin file
				//posicion = comienzo del espacio de bloques + id del bloque * tamaño del bloque
				position := int64(tempSuperblock.S_block_start + block*int32(binary.Size(Structs.Fileblock{})))
				if err := Utilities.ReadObject(file, &crrFileBlock, position); err != nil {
					return "", err
				}

				cleanContent := string(bytes.Trim(crrFileBlock.B_content[:], "\x00"))
				content += cleanContent

			} else {
				fmt.Print("indirectos")
			}
		}
		index++
	}

	fmt.Println("======End CONTENIDO DEL BLOQUE======")
	return content, nil
}

// stack
func pop(s *[]string) string {
	lastIndex := len(*s) - 1
	last := (*s)[lastIndex]
	*s = (*s)[:lastIndex]
	return last
}

func AppendToFileBlock(inode *Structs.Inode, newData string, file *os.File, superblock *Structs.Superblock) error {
	//vamos el ultimo bloque escrito
	index := 0
	for i := 0; i < 13; i++ {
		if inode.I_block[i] == -1 {
			index = i - 1
			break
		}
	}

	//ahora index contiene la posicion del ultimo bloque escrito
	var crrFileBlock Structs.Fileblock

	//leer el contenido del ultimo bloque escrito
	// offset = comienzo del espacio de bloques + id del bloque * tamaño del bloque
	offset := int64(superblock.S_block_start + inode.I_block[index]*int32(binary.Size(Structs.Fileblock{})))

	if err := Utilities.ReadObject(file, &crrFileBlock, offset); err != nil {
		return err
	}

	// Elimina bytes nulos y espacios en blanco al inicio y al final de cada línea
	cleanContent := string(bytes.Trim(crrFileBlock.B_content[:], "\x00"))
	cleanNewData := strings.TrimSpace(newData)

	// Elimina espacios internos y tabulaciones
	cleanContent = strings.ReplaceAll(cleanContent, " ", "")
	cleanContent = strings.ReplaceAll(cleanContent, "\t", "")
	cleanNewData = strings.ReplaceAll(cleanNewData, " ", "")
	cleanNewData = strings.ReplaceAll(cleanNewData, "\t", "")

	content := cleanContent + "\n" + cleanNewData

	//ahora que ya tenemos todo el contenido concatenado, vamos a escribirlo en el bloque
	//vamos a hacer otros bloques para meter lo que no cabe en el primero
	for i := 0; i < len(content); i += binary.Size(Structs.Fileblock{}) {
		var dataBlock Structs.Fileblock
		if i+binary.Size(Structs.Fileblock{}) > len(content) {
			copy(dataBlock.B_content[:], content[i:])
		} else {
			copy(dataBlock.B_content[0:64], content[i:i+binary.Size(Structs.Fileblock{})])
		}

		//sobreescribir el primer bloque
		//el offset es el inicio del primer bloque en la primera vuelta
		if i == 0 {
			if err := Utilities.WriteObject(file, dataBlock, offset); err != nil {
				return err
			}
			continue
		}

		//vamos a actualizar el offset a la siguiente vuelta.
		//offset va a ser el inicio del bloque libre
		offset = int64(superblock.S_first_blo)

		//escribir el bloque de datos en el archivo
		if err := Utilities.WriteObject(file, dataBlock, offset); err != nil {
			return err
		}

		superblock.S_first_blo += int32(binary.Size(dataBlock))

		//obtenemos la posicion del bit del bloque en el bitmap
		//possBitMap := superblock.S_blocks_count
		//superblock.S_blocks_count++
		//superblock.S_free_blocks_count--
		possBitMap := int32(offset-int64(superblock.S_block_start)) / int32(binary.Size(Structs.Fileblock{}))

		posicion := int64(superblock.S_bm_block_start + possBitMap)
		if err := Utilities.WriteObject(file, byte('1'), posicion); err != nil {
			return err
		}

		//actualizamos el inodo
		index++
		if inode.I_block[index] == -1 {
			superblock.S_free_blocks_count--
			superblock.S_blocks_count++
		}
		inode.I_block[index] = possBitMap

	}

	// Actualizar el tamaño del inodo
	inode.I_size = int32(len(content))

	//buscamos en donde es que queda el inodo

	posicion := int64(superblock.S_inode_start + int32(binary.Size(Structs.Inode{}))*inode.I_uid)
	if err := Utilities.WriteObject(file, *inode, posicion); err != nil {
		return fmt.Errorf("error al actualizar el inodo: %v", err)
	}

	//imprimimos el inodo
	Structs.PrintInode(*inode)

	existingData, err := GetInodeFileData(*inode, file, *superblock)
	if err != nil {
		return fmt.Errorf("error al leer el contenido del archivo: %v", err)
	}
	fmt.Println("existingData: ", existingData)

	return nil

}

func GetPossBlockFree(superblock Structs.Superblock, file *os.File) (int32, error) {
	var bitmap byte
	var i int32
	for i = 0; i < superblock.S_blocks_count; i++ {
		if err := Utilities.ReadObject(file, &bitmap, int64(superblock.S_bm_block_start+i)); err != nil {
			return -1, err
		}
		if bitmap == 0 {
			return i, nil
		}
	}
	return -1, errors.New("no hay bloques libres")
}

func CreateInodeByPath(path string, file *os.File, tempSuperblock Structs.Superblock) int32 {
	fmt.Println("======Start creando ruta ======")
	fmt.Println("path:", path)

	// split the path by /
	TempStepsPath := strings.Split(path, "/")
	StepsPath := TempStepsPath[1:] // remove the first diagonal xd

	fmt.Println("StepsPath:", StepsPath, "len(StepsPath):", len(StepsPath))
	for _, step := range StepsPath {
		fmt.Println("step:", step)
	}

	var Inode0 Structs.Inode
	// Read object from bin file
	//posicion = comienzo del espacio de inodos (primer inodo)
	if err := Utilities.ReadObject(file, &Inode0, int64(tempSuperblock.S_inode_start)); err != nil {
		return -1
	}
	fmt.Println("======End creando ruta======")

	//retornamos el inodo que se encuentra en la ruta
	return CreateInodeByPath_rec(StepsPath, Inode0, file, tempSuperblock)
}

func CreateInodeByPath_rec(StepsPath []string, Inode Structs.Inode, file *os.File, tempSuperblock Structs.Superblock) int32 {
	fmt.Println("======Start Creando path======")
	index := int32(0)
	SearchedName := strings.Replace(pop(&StepsPath), " ", "", -1)

	// Iterate over i_blocks from Inode
	for _, block := range Inode.I_block {
		if block != -1 {
			if index < 13 {
				//CASO DIRECTO

				var crrFolderBlock Structs.Folderblock
				// Read object from bin file
				if err := Utilities.ReadObject(file, &crrFolderBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(Structs.Folderblock{})))); err != nil {
					return -1
				}

				for _, folder := range crrFolderBlock.B_content {
					// fmt.Println("Folder found======")
					fmt.Println("Folder === Name:", string(folder.B_name[:]), "B_inodo", folder.B_inodo)

					if strings.Contains(string(folder.B_name[:]), SearchedName) {

						fmt.Println("len(StepsPath)", len(StepsPath), "StepsPath", StepsPath)
						if len(StepsPath) == 0 {
							fmt.Println("Folder found======")
							return folder.B_inodo
						} else {
							fmt.Println("NextInode======")
							var NextInode Structs.Inode
							// Read object from bin file
							if err := Utilities.ReadObject(file, &NextInode, int64(tempSuperblock.S_inode_start+folder.B_inodo*int32(binary.Size(Structs.Inode{})))); err != nil {
								return -1
							}
							return searchInodeByPath_rec(StepsPath, NextInode, file, tempSuperblock)
						}
					}
				}

			} else {
				fmt.Print("indirectos")
			}
		}
		index++
	}

	fmt.Println("======End BUSQUEDA INODO POR PATH======")
	return 0
}

// Imprimir inodos
func PrintInodes(file *os.File, sb Structs.Superblock) error {
	// Imprimir inodos
	//Structs.PrintSuperblock(sb)
	fmt.Println("\nInodos\n----------------")
	// Iterar sobre cada inodo
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := Structs.Inode{}
		// Deserializar el inodo
		err := Utilities.ReadObject(file, &inode, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil {
			return err
		}
		// Imprimir el inodo
		fmt.Printf("\nInodo %d en el byte %d \n", i, sb.S_inode_start+(i*sb.S_inode_size))
		Structs.PrintInode(inode)
	}
	return nil
}

// Impriir bloques
func PrintBlocks(file *os.File, sb Structs.Superblock) error {
	// Imprimir bloques
	Structs.PrintSuperblock(sb)
	fmt.Println("\nBloques\n----------------")
	// Iterar sobre cada inodo
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := Structs.Inode{}
		// Deserializar el inodo
		err := Utilities.ReadObject(file, &inode, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil {
			return err
		}
		// Iterar sobre cada bloque del inodo (apuntadores)
		for _, blockIndex := range inode.I_block {
			// Si el bloque no existe, salir
			if blockIndex == -1 {
				break
			}
			offset := int64(sb.S_block_start + blockIndex*sb.S_block_size)
			// Si el inodo es de tipo carpeta
			if inode.I_type[0] == '0' {
				block := Structs.Folderblock{}
				// Deserializar el bloque

				err := Utilities.ReadObject(file, &block, offset) // 64 porque es el tamaño de un bloque
				if err != nil {
					return err
				}
				// Imprimir el bloque
				//fmt.Printf("\nFolder block %d: en byte %d \n", blockIndex, int(offset))
				Structs.PrintFolderblock(block)

				// Si el inodo es de tipo archivo
			} else if inode.I_type[0] == '1' {
				block := Structs.Fileblock{}
				// Deserializar el bloque
				err := Utilities.ReadObject(file, &block, offset) // 64 porque es el tamaño de un bloque
				if err != nil {
					return err
				}
				// Imprimir el bloque
				//fmt.Printf("\narchivo block %d: en byte %d \n", blockIndex, offset)
				Structs.PrintFileblock(block)
				//block.Print()
			}

		}
	}

	return nil
}

func PrintBlocksInOrder(file *os.File, sb Structs.Superblock) error {
	// Imprimir bloques
	Structs.PrintSuperblock(sb)
	fmt.Println("\nBloques\n----------------")
	// Iterar sobre cada inodo
	fmt.Println("sb.S_blocks_count", sb.S_blocks_count)
	for i := int32(0); i < sb.S_blocks_count; i++ {
		fmt.Println("Bloque ", i)
		offset := int64(sb.S_block_start + i*sb.S_block_size)
		// Si el inodo es de tipo carpeta
		block := Structs.Folderblock{}
		// Deserializar el bloque

		err := Utilities.ReadObject(file, &block, offset)
		// 64 porque es el tamaño de un bloque
		if block.B_content[0].B_inodo > sb.S_inodes_count || err != nil {
			block := Structs.Fileblock{}
			// Deserializar el bloque
			err := Utilities.ReadObject(file, &block, offset) // 64 porque es el tamaño de un bloque
			if err != nil {
				return errors.New("el bloque no es ni de tipo archivo ni de tipo carpeta")
			}
			// Imprimir el bloque
			//fmt.Printf("\narchivo block %d: en byte %d \n", i, offset)
			Structs.PrintFileblock(block)
			continue
		}

		//imprimir el bloque
		Structs.PrintFolderblock(block)

	}

	return nil
}

func UpdateBitmapInode(file *os.File, sb *Structs.Superblock) error {
	// Actualizar el bitmap de inodos añadiendo 1 en la última posición disponible
	_, err := file.WriteAt([]byte{'1'}, int64(sb.S_bm_inode_start)+int64(sb.S_inodes_count))
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo: %v", err)
	}

	return nil
}

func UpdateBitmapBlock(file *os.File, sb *Structs.Superblock) error {

	// Actualizar el bitmap de bloques añadiendo 1 en la ultima posicion disponible
	_, err := file.WriteAt([]byte{'1'}, int64(sb.S_bm_block_start)+int64(sb.S_blocks_count))
	if err != nil {
		return err
	}

	return nil
}
