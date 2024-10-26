package Commands

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

func Fdisk(size int, path string, name string, unit string, type_ string, fit string, delete_ string, add int) error {
	fmt.Println("======Start FDISK======")
	fmt.Println("Size:", size)
	fmt.Println("Path:", path)
	fmt.Println("Name:", name)
	fmt.Println("Unit:", unit)
	fmt.Println("Type:", type_)
	fmt.Println("Fit:", fit)
	fmt.Println("Delete:", delete_)
	fmt.Println("Add:", add)

	if delete_ != "" {
		if path == "" || name == "" {
			fmt.Println("Error: Path and name are required")
			return errors.New("path and name are required")
		}
		//llaamada a la función DeletePartition
		err := DeletePartition(path, name, delete_)
		if err != nil {
			return err
		}
		return nil
	}

	if add != 0 {
		if path == "" || name == "" {
			fmt.Println("Error: Path and name are required")
			return errors.New("path and name are required")
		}
		//llamada a la función AddPartition
		lowercaseName := strings.ToLower(name)
		err := ModifyPartition(path, lowercaseName, add, unit)
		if err != nil {
			return err
		}
		return nil
	}

	err := validatefdisk(size, path, name, unit, type_, fit)
	if err != nil {
		return err
	}

	// Ajustar el tamaño en bytes
	if unit == "k" {
		size = size * 1024
	} else if unit == "m" {
		size = size * 1024 * 1024
	}

	// Abrir el archivo binario (disco) en la ruta proporcionada
	file, err := Utilities.OpenFile(path)
	if err != nil {
		fmt.Println("Error: Could not open file at path:", path)
		return errors.New("could not open file at path")
	}

	var TempMBR Structs.MRB
	// Leer el objeto desde el archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: Could not read MBR from file")
		return errors.New("could not read MBR from file")
	}

	// Imprimir el objeto MBR
	Structs.PrintMBR(TempMBR)

	fmt.Println("-------------")

	// Validaciones de las particiones
	var primaryCount, extendedCount, totalPartitions int
	var usedSpace int32 = 0

	for i := 0; i < 4; i++ {
		// comprobamos que el tamaño de la particion sea diferente a 0
		if TempMBR.Partitions[i].Size != 0 {
			totalPartitions++
			usedSpace += TempMBR.Partitions[i].Size

			if TempMBR.Partitions[i].Type[0] == 'p' {
				primaryCount++
			} else if TempMBR.Partitions[i].Type[0] == 'e' {
				extendedCount++
			}
		}
	}

	// Validar que no se exceda el número máximo de particiones primarias y extendidas
	if totalPartitions >= 4 {
		fmt.Println("Error: No se pueden crear más de 4 particiones primarias o extendidas en total.")
		return errors.New("no se pueden crear más de 4 particiones primarias o extendidas en total")
	}

	// Validar que solo haya una partición extendida
	if type_ == "e" && extendedCount > 0 {
		fmt.Println("Error: Solo se permite una partición extendida por disco.")
		return errors.New("solo se permite una partición extendida por disco")
	}

	// Validar que no se pueda crear una partición lógica sin una extendida
	if type_ == "l" && extendedCount == 0 {
		fmt.Println("Error: No se puede crear una partición lógica sin una partición extendida.")
		return errors.New("no se puede crear una partición lógica sin una partición extendida")
	}

	// Validar que el tamaño de la nueva partición no exceda el tamaño del disco
	if usedSpace+int32(size) > TempMBR.MbrSize {
		fmt.Println("Error: No hay suficiente espacio en el disco para crear esta partición.")
		return errors.New("no hay suficiente espacio en el disco para crear esta partición")
	}

	// Determinar la posición de inicio de la nueva partición
	var gap = int32(binary.Size(TempMBR))
	if totalPartitions > 0 {
		gap = TempMBR.Partitions[totalPartitions-1].Start + TempMBR.Partitions[totalPartitions-1].Size
	}

	// Encontrar una posición vacía para la nueva partición
	var nameBytes [16]byte
	for i := 0; i < 4; i++ {
		// Convertir el nombre a un arreglo de bytes de longitud fija
		copy(nameBytes[:], name)
		if bytes.Equal(TempMBR.Partitions[i].Name[:], nameBytes[:]) {
			fmt.Println("Error: Ya existe una partición con el nombre:", name)
			return errors.New("ya existe una partición con el nombre")
		}

		if TempMBR.Partitions[i].Size == 0 {
			if type_ == "p" || type_ == "e" {
				// Crear partición primaria o extendida
				TempMBR.Partitions[i].Size = int32(size)
				TempMBR.Partitions[i].Start = gap
				copy(TempMBR.Partitions[i].Name[:], name)
				copy(TempMBR.Partitions[i].Fit[:], fit)
				copy(TempMBR.Partitions[i].Status[:], "0")
				copy(TempMBR.Partitions[i].Type[:], type_)
				TempMBR.Partitions[i].Correlative = int32(totalPartitions + 1)

				if type_ == "e" {
					// Inicializar el primer EBR en la partición extendida
					ebrStart := gap // El primer EBR se coloca al inicio de la partición extendida
					ebr := Structs.EBR{
						PartFit:   fit[0],
						PartStart: ebrStart,
						PartSize:  0,
						PartNext:  -1,
					}
					copy(ebr.PartName[:], "")
					err := Utilities.WriteObject(file, ebr, int64(ebrStart))
					if err != nil {
						return err
					}
				}

				break
			}
		}
	}

	// Manejar la creación de particiones lógicas dentro de una partición extendida
	if type_ == "l" {
		for i := 0; i < 4; i++ {
			if TempMBR.Partitions[i].Type[0] == 'e' {
				ebrPos := TempMBR.Partitions[i].Start
				var ebr Structs.EBR
				for {
					err := Utilities.ReadObject(file, &ebr, int64(ebrPos))
					if err != nil {
						return err
					}

					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}

				// Calcular la posición de inicio de la nueva partición lógica
				newEBRPos := ebr.PartStart + ebr.PartSize                    // El nuevo EBR se coloca después de la partición lógica anterior
				logicalPartitionStart := newEBRPos + int32(binary.Size(ebr)) // El inicio de la partición lógica es justo después del EBR

				// Ajustar el siguiente EBR
				ebr.PartNext = newEBRPos
				err := Utilities.WriteObject(file, ebr, int64(ebrPos))
				if err != nil {
					return err
				}

				// Crear y escribir el nuevo EBR
				newEBR := Structs.EBR{
					PartFit:   fit[0],
					PartStart: logicalPartitionStart,
					PartSize:  int32(size),
					PartNext:  -1,
				}
				copy(newEBR.PartName[:], name)
				err = Utilities.WriteObject(file, newEBR, int64(newEBRPos))
				if err != nil {
					return err
				}

				// Imprimir el nuevo EBR creado
				fmt.Println("Nuevo EBR creado:")
				Structs.PrintEBR(newEBR)
				fmt.Println("")

				// Imprimir todos los EBRs en la partición extendida
				fmt.Println("Imprimiendo todos los EBRs en la partición extendida:")
				ebrPos = TempMBR.Partitions[i].Start
				for {
					err := Utilities.ReadObject(file, &ebr, int64(ebrPos))
					if err != nil {
						fmt.Println("Error al leer EBR:", err)
						break
					}
					Structs.PrintEBR(ebr)
					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}

				break
			}
		}
		fmt.Println("")
	}

	// Sobrescribir el MBR
	if err := Utilities.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println("Error: Could not write MBR to file")
		return errors.New("could not write MBR to file")
	}

	var TempMBR2 Structs.MRB
	// Leer el objeto nuevamente para verificar
	if err := Utilities.ReadObject(file, &TempMBR2, 0); err != nil {
		fmt.Println("Error: Could not read MBR from file after writing")
		return errors.New("could not read MBR from file after writing")
	}

	// Imprimir el objeto MBR actualizado
	Structs.PrintMBR(TempMBR2)

	// Cerrar el archivo binario
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error: Could not close file:", err)
		}
	}(file)

	fmt.Println("======FIN FDISK======")
	fmt.Println("")

	return nil
}

// DeletePartition Función para eliminar particiones
func DeletePartition(path string, name string, delete_ string) error {
	fmt.Println("======Start DELETE PARTITION======")
	fmt.Println("Path:", path)
	fmt.Println("Name:", name)
	fmt.Println("Delete type:", delete_)

	// Abrir el archivo binario en la ruta proporcionada
	file, err := Utilities.OpenFile(path)
	if err != nil {
		fmt.Println("Error: Could not open file at path:", path)
		return errors.New("could not open file at path" + path)
	}

	var TempMBR Structs.MRB
	// Leer el objeto desde el archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: Could not read MBR from file")
		return errors.New("could not read MBR from file")
	}

	// Buscar la partición por nombre
	found := false
	for i := 0; i < 4; i++ {
		// Limpiar los caracteres nulos al final del nombre de la partición
		partitionName := strings.TrimRight(string(TempMBR.Partitions[i].Name[:]), "\x00")
		if partitionName == name {
			found = true

			// Si es una partición extendida, eliminar las particiones lógicas dentro de ella
			if TempMBR.Partitions[i].Type[0] == 'e' {
				fmt.Println("Eliminando particiones lógicas dentro de la partición extendida...")
				ebrPos := TempMBR.Partitions[i].Start
				var ebr Structs.EBR
				for {
					err := Utilities.ReadObject(file, &ebr, int64(ebrPos))
					if err != nil {
						fmt.Println("Error al leer EBR:", err)
						break
					}
					// Detener el bucle si el EBR está vacío
					if ebr.PartStart == 0 && ebr.PartSize == 0 {
						fmt.Println("EBR vacío encontrado, deteniendo la búsqueda.")
						break
					}
					// Depuración: Mostrar el EBR leído
					fmt.Println("EBR leído antes de eliminar:")
					Structs.PrintEBR(ebr)

					// Eliminar partición lógica
					if delete_ == "fast" {
						ebr = Structs.EBR{} // Resetear el EBR manualmente
						err := Utilities.WriteObject(file, ebr, int64(ebrPos))
						if err != nil {
							return errors.New("could not write EBR to file (1)" + err.Error())
						} // Sobrescribir el EBR reseteado
					} else if delete_ == "full" {
						err := Utilities.FillWithZeros(file, ebr.PartStart, ebr.PartSize)
						if err != nil {
							return err
						}
						ebr = Structs.EBR{} // Resetear el EBR manualmente
						err = Utilities.WriteObject(file, ebr, int64(ebrPos))
						if err != nil {
							return errors.New("could not write EBR to file (2)" + err.Error())
						} // Sobrescribir el EBR reseteado
					}

					// Depuración: Mostrar el EBR después de eliminar
					fmt.Println("EBR después de eliminar:")
					Structs.PrintEBR(ebr)

					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}
			}

			// Proceder a eliminar la partición (extendida, primaria o lógica)
			if delete_ == "fast" {
				// Eliminar rápido: Resetear manualmente los campos de la partición
				TempMBR.Partitions[i] = Structs.Partition{} // Resetear la partición manualmente
				fmt.Println("Partición eliminada en modo Fast.")
			} else if delete_ == "full" {
				// Eliminar completamente: Resetear manualmente y sobrescribir con '\0'
				start := TempMBR.Partitions[i].Start
				size := TempMBR.Partitions[i].Size
				TempMBR.Partitions[i] = Structs.Partition{} // Resetear la partición manualmente
				// Escribir '\0' en el espacio de la partición en el disco
				err := Utilities.FillWithZeros(file, start, size)
				if err != nil {
					return err
				}
				fmt.Println("Partición eliminada en modo Full.")

				// Leer y verificar si el área está llena de ceros
				Utilities.VerifyZeros(file, start, size)
			}
			break
		}
	}

	if !found {
		// Buscar particiones lógicas si no se encontró en el MBR
		fmt.Println("Buscando en particiones lógicas dentro de las extendidas...")
		for i := 0; i < 4; i++ {
			if TempMBR.Partitions[i].Type[0] == 'e' { // Solo buscar dentro de particiones extendidas
				ebrPos := TempMBR.Partitions[i].Start
				var ebr Structs.EBR
				for {
					err := Utilities.ReadObject(file, &ebr, int64(ebrPos))
					if err != nil {
						fmt.Println("Error al leer EBR:", err)
						break
					}

					// Depuración: Mostrar el EBR leído
					fmt.Println("EBR leído:")
					Structs.PrintEBR(ebr)

					logicalName := strings.TrimRight(string(ebr.PartName[:]), "\x00")
					if logicalName == name {
						found = true
						// Eliminar la partición lógica
						if delete_ == "fast" {
							ebr = Structs.EBR{} // Resetear el EBR manualmente
							err = Utilities.WriteObject(file, ebr, int64(ebrPos))
							if err != nil {
								return err
							} // Sobrescribir el EBR reseteado
							fmt.Println("Partición lógica eliminada en modo Fast.")
						} else if delete_ == "full" {
							err := Utilities.FillWithZeros(file, ebr.PartStart, ebr.PartSize)
							if err != nil {
								return err
							}
							ebr = Structs.EBR{} // Resetear el EBR manualmente
							err = Utilities.WriteObject(file, ebr, int64(ebrPos))
							if err != nil {
								return err
							} // Sobrescribir el EBR reseteado
							Utilities.VerifyZeros(file, ebr.PartStart, ebr.PartSize)
							fmt.Println("Partición lógica eliminada en modo Full.")
						}
						break
					}

					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}
			}
			if found {
				break
			}
		}
	}

	if !found {
		fmt.Println("Error: No se encontró la partición con el nombre:", name)
		return errors.New("no se encontró la partición con el nombre")
	}

	// Sobrescribir el MBR
	if err := Utilities.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println("Error: Could not write MBR to file")
		return errors.New("could not write MBR to file")
	}

	// Leer el MBR actualizado y mostrarlo
	fmt.Println("MBR actualizado después de la eliminación:")
	Structs.PrintMBR(TempMBR)

	// Si es una partición extendida, mostrar los EBRs actualizados
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Type[0] == 'e' {
			fmt.Println("Imprimiendo EBRs actualizados en la partición extendida:")
			ebrPos := TempMBR.Partitions[i].Start
			var ebr Structs.EBR
			for {
				err := Utilities.ReadObject(file, &ebr, int64(ebrPos))
				if err != nil {
					fmt.Println("Error al leer EBR:", err)
					break
				}
				// Detener el bucle si el EBR está vacío
				if ebr.PartStart == 0 && ebr.PartSize == 0 {
					fmt.Println("EBR vacío encontrado, deteniendo la búsqueda.")
					break
				}
				// Depuración: Imprimir cada EBR leído
				fmt.Println("EBR leído después de actualización:")
				Structs.PrintEBR(ebr)
				if ebr.PartNext == -1 {
					break
				}
				ebrPos = ebr.PartNext
			}
		}
	}

	// Cerrar el archivo binario
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error: Could not close file:", err)
		}
	}(file)

	fmt.Println("======FIN DELETE PARTITION======")
	return nil
}

func ModifyPartition(path string, name string, add int, unit string) error {
	fmt.Println("======Start MODIFY PARTITION======")
	// Abrir el archivo binario en la ruta proporcionada
	file, err := Utilities.OpenFile(path)
	if err != nil {
		fmt.Println("Error: Could not open file at path:", path)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error: Could not close file:", err)
		}
	}(file)

	// Leer el MBR
	var TempMBR Structs.MRB
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: Could not read MBR from file")
		return err
	}

	// Imprimir MBR antes de modificar
	fmt.Println("MBR antes de la modificación:")
	Structs.PrintMBR(TempMBR)

	// Buscar la partición por nombre
	var foundPartition *Structs.Partition
	var partitionType byte

	// Revisar si la partición es primaria o extendida
	for i := 0; i < 4; i++ {
		partitionName := strings.TrimRight(string(TempMBR.Partitions[i].Name[:]), "\x00")
		if partitionName == name {
			foundPartition = &TempMBR.Partitions[i]
			partitionType = TempMBR.Partitions[i].Type[0]
			//vamos a ver si hay particiones que le sigan
			if (i + 1) < 4 {
				partitionNamePosterior := strings.TrimRight(string(TempMBR.Partitions[i+1].Name[:]), "\x00")
				if partitionNamePosterior != "" && add > 0 {
					//verificamos si hay espacio para que quepa la partición
					if TempMBR.Partitions[i].Size+int32(add)+TempMBR.Partitions[i].Start > TempMBR.Partitions[i+1].Start {
						fmt.Println("Error: No hay suficiente espacio para aumentar la partición")
						return errors.New("Error: No hay suficiente espacio para aumentar la partición")
					}
				}
				if add < 0 {
					//verificamos que el size no quede negativo
					if TempMBR.Partitions[i].Size+int32(add) < 0 {
						fmt.Println("Error: No es posible reducir la partición por debajo de 0")
						return errors.New("Error: No es posible reducir la partición por debajo de 0")
					}
				}
			}
			break
		}
	}

	// Si no se encuentra en las primarias/extendidas, buscar en las particiones lógicas
	if foundPartition == nil {
		for i := 0; i < 4; i++ {
			if TempMBR.Partitions[i].Type[0] == 'e' {
				ebrPos := TempMBR.Partitions[i].Start
				var ebr Structs.EBR
				for {
					if err := Utilities.ReadObject(file, &ebr, int64(ebrPos)); err != nil {
						fmt.Println("Error al leer EBR:", err)
						return err
					}

					ebrName := strings.TrimRight(string(ebr.PartName[:]), "\x00")
					if ebrName == name {
						partitionType = 'l' // Partición lógica
						foundPartition = &Structs.Partition{
							Start: ebr.PartStart,
							Size:  ebr.PartSize,
						}
						if ebr.PartNext != -1 && add > 0 {
							//verificamos si hay espacio para que quepa la partición
							if ebr.PartSize+int32(add) > ebr.PartNext-ebr.PartStart {
								fmt.Println("Error: No hay suficiente espacio para aumentar la partición")
								return errors.New("Error: No hay suficiente espacio para aumentar la partición")
							}
						}
						if add < 0 {
							//verificamos que el size no quede negativo
							if ebr.PartSize+int32(add) < 0 {
								fmt.Println("Error: No es posible reducir la partición por debajo de 0")
								return errors.New("Error: No es posible reducir la partición por debajo de 0")
							}
						}
						break
					}

					// Continuar buscando el siguiente EBR
					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}
				if foundPartition != nil {
					break
				}
			}
		}
	}

	// Verificar si la partición fue encontrada
	if foundPartition == nil {
		fmt.Println("Error: No se encontró la partición con el nombre:", name)
		return nil // Salir si no se encuentra la partición
	}

	// Convertir unidades a bytes
	var addBytes int
	if unit == "k" {
		addBytes = add * 1024
	} else if unit == "m" {
		addBytes = add * 1024 * 1024
	} else {
		fmt.Println("Error: Unidad desconocida, debe ser 'k' o 'm'")
		return nil // Salir si la unidad no es válida
	}

	// Flag para saber si continuar o no
	var shouldModify = true

	// Comprobar si es posible agregar o quitar espacio
	//if add > 0 {
	//	// Agregar espacio: verificar si hay suficiente espacio libre después de la partición
	//	nextPartitionStart := foundPartition.Start + foundPartition.Size
	//	if partitionType == 'l' {
	//		// Para particiones lógicas, verificar con el siguiente EBR o el final de la partición extendida
	//		for i := 0; i < 4; i++ {
	//			if TempMBR.Partitions[i].Type[0] == 'e' {
	//				extendedPartitionEnd := TempMBR.Partitions[i].Start + TempMBR.Partitions[i].Size
	//				if nextPartitionStart+int32(addBytes) > extendedPartitionEnd {
	//					fmt.Println("Error: No hay suficiente espacio libre dentro de la partición extendida")
	//					shouldModify = false
	//				}
	//				break
	//			}
	//		}
	//	} else {
	//		// Para primarias o extendidas
	//		if nextPartitionStart+int32(addBytes) > TempMBR.MbrSize {
	//			fmt.Println("Error: No hay suficiente espacio libre después de la partición")
	//			shouldModify = false
	//		}
	//	}
	//} else {
	// Quitar espacio: verificar que no se reduzca el tamaño por debajo de 0
	if foundPartition.Size+int32(addBytes) < 0 {
		fmt.Println("Error: No es posible reducir la partición por debajo de 0")
		shouldModify = false
	}
	//}

	// Solo modificar si no hay errores
	if shouldModify {
		foundPartition.Size += int32(addBytes)
	} else {
		fmt.Println("No se realizaron modificaciones debido a un error.")
		return nil // Salir si hubo un error
	}

	// Si es una partición lógica, sobrescribir el EBR
	if partitionType == 'l' {
		ebrPos := foundPartition.Start
		var ebr Structs.EBR
		if err := Utilities.ReadObject(file, &ebr, int64(ebrPos)); err != nil {
			fmt.Println("Error al leer EBR:", err)
			return err
		}

		// Actualizar el tamaño en el EBR y escribirlo de nuevo
		ebr.PartSize = foundPartition.Size
		if err := Utilities.WriteObject(file, ebr, int64(ebrPos)); err != nil {
			fmt.Println("Error al escribir el EBR actualizado:", err)
			return err
		}

		// Imprimir el EBR modificado
		fmt.Println("EBR modificado:")
		Structs.PrintEBR(ebr)
	}

	// Sobrescribir el MBR actualizado
	if err := Utilities.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println("Error al escribir el MBR actualizado:", err)
		return err
	}

	// Imprimir el MBR modificado
	fmt.Println("MBR después de la modificación:")
	Structs.PrintMBR(TempMBR)

	fmt.Println("======END MODIFY PARTITION======")
	return nil
}

func validatefdisk(size int, path string, name string, unit string, type_ string, fit string) error {

	if size <= 0 {
		fmt.Println("Error: Size must be greater than 0")
		return errors.New("size must be greater than 0")
	}

	// Validar fit (b/w/f)
	if unit != "b" && unit != "k" && unit != "m" {
		fmt.Println("Error: Unit must be 'b', 'k' or 'm'")
		return errors.New("unit must be 'b', 'k' or 'm'")
	}

	if path == "" {
		fmt.Println("Error: Path is required")
		return errors.New("path is required")
	}

	if type_ != "p" && type_ != "e" && type_ != "l" {
		fmt.Println("Error: Type must be 'p', 'e', or 'l'")
		return errors.New("type must be 'p', 'e', or 'l'")
	}

	if fit != "bf" && fit != "ff" && fit != "wf" {
		fmt.Println("Error: Fit must be 'bf', 'ff', or 'wf'")
		return errors.New("fit must be 'bf', 'ff', or 'wf'")
	}

	if name == "" {
		fmt.Println("Error: Name is required")
		return errors.New("name is required")
	}

	return nil
}
