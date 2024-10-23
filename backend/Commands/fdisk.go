package Commands

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"proyecto1/Structs"
	"proyecto1/Utilities"
)

func Fdisk(size int, path string, name string, unit string, type_ string, fit string) error {
	fmt.Println("======Start FDISK======")
	fmt.Println("Size:", size)
	fmt.Println("Path:", path)
	fmt.Println("Name:", name)
	fmt.Println("Unit:", unit)
	fmt.Println("Type:", type_)
	fmt.Println("Fit:", fit)

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
