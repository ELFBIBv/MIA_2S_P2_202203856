package Commands

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"proyecto1/DiskManagement"
	"proyecto1/Structs"
	"proyecto1/Utilities"
	"strings"
)

func RepMBR(id string, path string) error {
	var ParticionesMontadas DiskManagement.MountedPartition
	var ParticionEncontrada bool

	for _, Particiones := range DiskManagement.GetMountedPartitions() {
		for _, Particion := range Particiones {
			if Particion.ID == id {
				ParticionesMontadas = Particion
				ParticionEncontrada = true
				break
			}
		}
		if ParticionEncontrada {
			break
		}
	}

	if !ParticionEncontrada {
		//Consola.AddString("Error REP MBR: No se encontró la partición con el ID: %s.\n" + id)
		fmt.Printf("Error REP MBR: No se encontró la partición con el ID: %s.\n", id)
		return fmt.Errorf("No se encontró la partición con el ID: %s.", id)
	}

	archivo, err := Utilities.OpenFile(ParticionesMontadas.Path)
	if err != nil {
		return err
	}
	defer func(archivo *os.File) {
		err := archivo.Close()
		if err != nil {
			fmt.Println("Error REP MBR: Error al cerrar el archivo.")
		}
	}(archivo)

	var MBRTemporal Structs.MRB
	if err := Utilities.ReadObject(archivo, &MBRTemporal, 0); err != nil {
		return err
	}

	dot := "digraph G {\n"
	dot += "node [shape=plaintext];\n"
	dot += "fontname=\"Courier New\";\n"
	dot += "title [label=\"Reporte MBR\"];\n"
	dot += "mbrTable [label=<\n"
	dot += "<table border='1' cellborder='1' cellspacing='0'>\n"
	dot += "<tr><td bgcolor=\"blue\" colspan='2'>MBR</td></tr>\n"
	dot += fmt.Sprintf("<tr><td>Tamaño</td><td>%d</td></tr>\n", MBRTemporal.MbrSize)
	dot += fmt.Sprintf("<tr><td>Fecha De Creación</td><td>%s</td></tr>\n", string(MBRTemporal.CreationDate[:]))
	dot += fmt.Sprintf("<tr><td>Ajuste</td><td>%s</td></tr>\n", string(MBRTemporal.Fit[:]))
	dot += fmt.Sprintf("<tr><td>Signature</td><td>%d</td></tr>\n", MBRTemporal.Signature)
	dot += "</table>\n"
	dot += ">];\n"

	for i, Particion := range MBRTemporal.Partitions {
		if Particion.Size != 0 {
			dot += fmt.Sprintf("PA%d [label=<\n", i+1)
			dot += "<table border='1' cellborder='1' cellspacing='0'>\n"
			dot += fmt.Sprintf("<tr><td bgcolor=\"red\" colspan='2'>Partición %d</td></tr>\n", i+1)
			dot += fmt.Sprintf("<tr><td>Estado</td><td>%s</td></tr>\n", string(Particion.Status[:]))
			dot += fmt.Sprintf("<tr><td>Tipo</td><td>%s</td></tr>\n", string(Particion.Type[:]))
			dot += fmt.Sprintf("<tr><td>Ajuste</td><td>%s</td></tr>\n", string(Particion.Fit[:]))
			dot += fmt.Sprintf("<tr><td>Incio</td><td>%d</td></tr>\n", Particion.Start)
			dot += fmt.Sprintf("<tr><td>Tamaño</td><td>%d</td></tr>\n", Particion.Size)
			dot += fmt.Sprintf("<tr><td>Nombre</td><td>%s</td></tr>\n", strings.Trim(string(Particion.Name[:]), "\x00"))
			dot += fmt.Sprintf("<tr><td>Correlativo</td><td>%d</td></tr>\n", Particion.Correlative)
			dot += "</table>\n"
			dot += ">];\n"
			if Particion.Type[0] == 'e' {
				var EBR Structs.EBR
				if err := Utilities.ReadObject(archivo, &EBR, int64(Particion.Start)); err != nil {
					return err
				}
				if EBR.PartSize != 0 {
					var ContadorLogicas = 0
					dot += "subgraph cluster_0 {style=filled;color=lightgrey;label = \"Partición Extendida\";"
					dot += "fontname=\"Courier New\";"
					for {
						dot += fmt.Sprintf("EBR%d [label=<\n", EBR.PartStart)
						dot += "<table border='1' cellborder='1' cellspacing='0'>\n"
						dot += "<tr><td bgcolor=\"green\" colspan='2'>EBR</td></tr>\n"
						dot += fmt.Sprintf("<tr><td>Nombre</td><td>%s</td></tr>\n", strings.Trim(string(EBR.PartName[:]), "\x00"))
						dot += fmt.Sprintf("<tr><td>Ajuste</td><td>%s</td></tr>\n", string(EBR.PartFit))
						dot += fmt.Sprintf("<tr><td>Inicio</td><td>%d</td></tr>\n", EBR.PartStart)
						dot += fmt.Sprintf("<tr><td>Tamaño</td><td>%d</td></tr>\n", EBR.PartSize)
						dot += fmt.Sprintf("<tr><td>Siguiente</td><td>%d</td></tr>\n", EBR.PartNext)
						dot += "</table>\n"
						dot += ">];\n"

						dot += fmt.Sprintf("Pl%d [label=<\n", ContadorLogicas)
						dot += "<table border='1' cellborder='1' cellspacing='0'>\n"
						dot += "<tr><td bgcolor=\"purple\" colspan='2'>Partición Lógica</td></tr>\n"
						dot += fmt.Sprintf("<tr><td>Estado</td><td>%s</td></tr>\n", "0")
						dot += fmt.Sprintf("<tr><td>Tipo</td><td>%s</td></tr>\n", "l")
						dot += fmt.Sprintf("<tr><td>Ajuste</td><td>%s</td></tr>\n", string(EBR.PartFit))
						dot += fmt.Sprintf("<tr><td>Incio</td><td>%d</td></tr>\n", EBR.PartStart)
						dot += fmt.Sprintf("<tr><td>Tamaño</td><td>%d</td></tr>\n", EBR.PartSize)
						dot += fmt.Sprintf("<tr><td>Nombre</td><td>%s</td></tr>\n", strings.Trim(string(EBR.PartName[:]), "\x00"))
						dot += fmt.Sprintf("<tr><td>Correlativo</td><td>%d</td></tr>\n", ContadorLogicas+1)
						dot += "</table>\n"
						dot += ">];\n"
						if EBR.PartNext == -1 {
							break
						}
						if err := Utilities.ReadObject(archivo, &EBR, int64(EBR.PartNext)); err != nil {
							fmt.Print("Error al leer siguiente EBR: %v\n" + string(err.Error()))
							return err
						}
						ContadorLogicas++
					}
					dot += "}\n"
				}
			}
		}
	}
	dot += "}\n"
	dotFilePath := "ReporteMBR.dot"
	err = os.WriteFile(dotFilePath, []byte(dot), 0644)
	if err != nil {
		fmt.Println("Error REP MBR:" + string(err.Error()))
		return errors.New("Error REP MBR:" + string(err.Error()))
	}
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("Error REP MBR:" + string(err.Error()))
			return errors.New("Error REP MBR:" + string(err.Error()))
		}
	}
	cmd := exec.Command("dot", "-Tjpg", dotFilePath, "-o", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error REP MBR:" + string(err.Error()))
		return errors.New("Error REP MBR:" + string(err.Error()))
	}
	return nil
}

func RepDisk(id string, path string) error {
	var ParticionesMontadas DiskManagement.MountedPartition
	var ParticionEncontrada bool

	for _, Particiones := range DiskManagement.GetMountedPartitions() {
		for _, Particion := range Particiones {
			if Particion.ID == id {
				ParticionesMontadas = Particion
				ParticionEncontrada = true
				break
			}
		}
		if ParticionEncontrada {
			break
		}
	}

	if !ParticionEncontrada {
		fmt.Printf("Error REP DISK: No se encontró la partición con el ID: %s.\n", id)
		return fmt.Errorf("No se encontró la partición con el ID: %s.", id)

	}

	archivo, err := Utilities.OpenFile(ParticionesMontadas.Path)
	if err != nil {
		fmt.Println("Error REP DISK:" + string(err.Error()))
		return err
	}
	defer func(archivo *os.File) {
		err := archivo.Close()
		if err != nil {
			fmt.Println("Error REP DISK: Error al cerrar el archivo.")
		}
	}(archivo)

	var MBRTemporal Structs.MRB
	if err := Utilities.ReadObject(archivo, &MBRTemporal, 0); err != nil {
		fmt.Println("Error REP DISK:" + string(err.Error()))
		return err
	}

	TamanoTotal := float64(MBRTemporal.MbrSize)
	EspacioUsado := 0.0

	dot := "digraph G {\n"
	dot += "labelloc=\"t\"\n"
	dot += "node [shape=plaintext];\n"
	dot += "fontname=\"Courier New\";\n"
	dot += "title [label=\"Reporte Disk\"];\n"
	dot += "subgraph cluster1 {\n"
	dot += "fontname=\"Courier New\";\n"
	dot += "label=\"\"\n"
	dot += "disco [shape=none label=<\n"
	dot += "<TABLE border=\"0\" cellspacing=\"4\" cellpadding=\"5\" color=\"skyblue\">\n"
	dot += "<TR><TD bgcolor=\"#a7d0d2\" border=\"1\" cellpadding=\"65\">MBR</TD>\n"

	for i, Particion := range MBRTemporal.Partitions {
		if Particion.Status[0] != 0 {
			Size := float64(Particion.Size)
			EspacioUsado += Size

			if Particion.Type[0] == 'e' || Particion.Type[0] == 'E' {
				dot += "<TD border=\"1\" width=\"75\">\n"
				dot += "<TABLE border=\"0\" cellspacing=\"4\" cellpadding=\"10\">\n"
				dot += "<TR><TD bgcolor=\"skyblue\" border=\"1\" colspan=\"5\" height=\"75\"> Partición Extendida<br/></TD></TR>\n"

				EspacioLibreExtendida := Size
				finEbr := Particion.Start

				var EBR Structs.EBR
				if err := Utilities.ReadObject(archivo, &EBR, int64(Particion.Start)); err != nil {
					fmt.Println("Error al leer EBR:" + string(err.Error()))
					return err
				}
				if EBR.PartSize != 0 {
					for {
						var ebr Structs.EBR
						if err := Utilities.ReadObject(archivo, &ebr, int64(finEbr)); err != nil {
							fmt.Println("Error al leer EBR:" + string(err.Error()))
							return err
						}

						TamanoEBR := float64(ebr.PartSize)
						EspacioUsado += TamanoEBR
						EspacioLibreExtendida -= TamanoEBR

						dot += "<TR>\n"
						dot += "<TD bgcolor=\"#264b5e\" border=\"1\" height=\"185\">EBR</TD>\n"
						dot += fmt.Sprintf("<TD bgcolor=\"#546eab\" border=\"1\" cellpadding=\"25\">Partición Lógica<br/>%.2f%% del Disco</TD>\n", (TamanoEBR/TamanoTotal)*100)
						dot += "</TR>\n"
						if ebr.PartNext <= 0 {
							break
						}
						finEbr = ebr.PartNext
					}
				}
				dot += "<TR>\n"
				dot += fmt.Sprintf("<TD bgcolor=\"#f1e6d2\" border=\"1\" colspan=\"5\"> Espacio Libre Dentro De La Partición Extendida<br/>%.2f%% del Disco</TD>\n", (EspacioLibreExtendida/TamanoTotal)*100)
				dot += "</TR>\n"

				dot += "</TABLE>\n</TD>\n"
			} else if Particion.Type[0] == 'p' || Particion.Type[0] == 'P' {
				dot += fmt.Sprintf("<TD bgcolor=\"#4697b4\" border=\"1\" cellpadding=\"20\">Partición Primaria %d<br/>%.2f%% del Disco</TD>\n", i+1, (Size/TamanoTotal)*100)
			}
		}
	}

	Porcentaje := 100.0
	for _, partition := range MBRTemporal.Partitions {
		if partition.Status[0] != 0 {
			Size := float64(partition.Size)
			Porcentaje -= (Size / TamanoTotal) * 100
		}
	}

	dot += fmt.Sprintf("<TD bgcolor=\"#f1e6d2\" border=\"1\" cellpadding=\"20\">Espacio Libre<br/>%.2f%% del Disco</TD>\n", Porcentaje)
	dot += "</TR>\n"
	dot += "</TABLE>\n"
	dot += ">];\n"
	dot += "}\n"
	dot += "}\n"

	RutaReporte := "ReporteDisk.dot"
	err = os.WriteFile(RutaReporte, []byte(dot), 0644)
	if err != nil {
		fmt.Println("Error REP DISK:" + string(err.Error()))
		return fmt.Errorf("Error REP DISK:" + string(err.Error()))
	}
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("Error REP DISK:" + string(err.Error()))
			return fmt.Errorf("Error REP DISK:" + string(err.Error()))
		}
	}
	cmd := exec.Command("dot", "-Tjpg", RutaReporte, "-o", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error REP DISK:" + string(err.Error()))
		return fmt.Errorf("Error REP DISK:" + string(err.Error()))
	}
	return nil
}

func RepSB(id string, path string) error {
	var ParticionesMontadas DiskManagement.MountedPartition
	var ParticionEncontrada bool

	for _, Particiones := range DiskManagement.GetMountedPartitions() {
		for _, Particion := range Particiones {
			if Particion.ID == id {
				ParticionesMontadas = Particion
				ParticionEncontrada = true
				break
			}
		}
		if ParticionEncontrada {
			break
		}
	}

	if !ParticionEncontrada {
		fmt.Printf("Error REP SB: No se encontró la partición con el ID: %s.\n", id)
		return fmt.Errorf("No se encontró la partición con el ID: %s.", id)
	}

	archivo, err := Utilities.OpenFile(ParticionesMontadas.Path)
	if err != nil {
		fmt.Println("Error REP SB:" + string(err.Error()))
		return err
	}
	defer func(archivo *os.File) {
		err := archivo.Close()
		if err != nil {
			fmt.Println("Error REP SB: Error al cerrar el archivo.")
		}
	}(archivo)

	var MBRTemporal Structs.MRB
	if err := Utilities.ReadObject(archivo, &MBRTemporal, 0); err != nil {
		fmt.Println("Error REP SB:" + string(err.Error()))
		return err
	}

	var index = -1
	for i := 0; i < 4; i++ {
		if MBRTemporal.Partitions[i].Size != 0 {
			if strings.Contains(string(MBRTemporal.Partitions[i].Id[:]), id) {
				if MBRTemporal.Partitions[i].Status[0] == '1' {
					index = i
				} else {
					fmt.Println("Error REP SB: La partición con el ID:%s no está montada.\n" + id)
					return errors.New("La partición con el ID:%s no está montada.\n" + id)
				}
				break
			}
		}
	}

	if index == -1 {
		fmt.Printf("Error REP SB: No se encontró la partición con el ID: %s.\n", id)
		return fmt.Errorf("No se encontró la partición con el ID: %s.", id)
	}

	var TemporalSuperBloque = Structs.Superblock{}
	if err := Utilities.ReadObject(archivo, &TemporalSuperBloque, int64(MBRTemporal.Partitions[index].Start)); err != nil {
		fmt.Println("Error REP SB: Error al leer el SuperBloque.")
		return err
	}

	dot := "digraph G {\n"
	dot += "node [shape=plaintext];\n"
	dot += "fontname=\"Courier New\";\n"
	dot += "title [label=\"Reporte SuperBloque\"];\n"
	dot += "SBTable [label=<\n"
	dot += "<table border='1' cellborder='1' cellspacing='0'>\n"
	dot += "<tr><td bgcolor=\"skyblue\" colspan='2'>Super Bloque</td></tr>\n"
	dot += fmt.Sprintf("<tr><td>SB FileSystem Type</td><td>%d</td></tr>\n", int(TemporalSuperBloque.S_filesystem_type))
	dot += fmt.Sprintf("<tr><td>SB Inodes Count</td><td>%d</td></tr>\n", int(TemporalSuperBloque.S_inodes_count))
	dot += fmt.Sprintf("<tr><td>SB Blocks Count</td><td>%d</td></tr>\n", int(TemporalSuperBloque.S_blocks_count))
	dot += fmt.Sprintf("<tr><td>SB Free Blocks Count</td><td>%d</td></tr>\n", int(TemporalSuperBloque.S_free_blocks_count))
	dot += fmt.Sprintf("<tr><td>SB Free Inodes Count</td><td>%d</td></tr>\n", int(TemporalSuperBloque.S_free_inodes_count))
	dot += fmt.Sprintf("<tr><td>SB Mtime</td><td>%s</td></tr>\n", string(TemporalSuperBloque.S_mtime[:]))
	dot += fmt.Sprintf("<tr><td>SB Umtime</td><td>%s</td></tr>\n", string(TemporalSuperBloque.S_umtime[:]))
	dot += fmt.Sprintf("<tr><td>SB Mnt Count</td><td>%d</td></tr>\n", int(TemporalSuperBloque.S_mnt_count))
	dot += fmt.Sprintf("<tr><td>SB Magic</td><td>%d</td></tr>\n", int(TemporalSuperBloque.S_magic))
	dot += fmt.Sprintf("<tr><td>SB Inode Size</td><td>%d</td></tr>\n", int(TemporalSuperBloque.S_inode_size))
	dot += fmt.Sprintf("<tr><td>SB Block Size</td><td>%d</td></tr>\n", int(TemporalSuperBloque.S_block_size))
	dot += fmt.Sprintf("<tr><td>SB Fist Inode</td><td>%d</td></tr>\n", int(TemporalSuperBloque.S_first_ino))
	dot += fmt.Sprintf("<tr><td>SB First Block</td><td>%d</td></tr>\n", int(TemporalSuperBloque.S_first_blo))
	dot += fmt.Sprintf("<tr><td>SB Bm Inode Start</td><td>%d</td></tr>\n", int(TemporalSuperBloque.S_bm_inode_start))
	dot += fmt.Sprintf("<tr><td>SB Bm Block Start</td><td>%d</td></tr>\n", int(TemporalSuperBloque.S_bm_block_start))
	dot += fmt.Sprintf("<tr><td>SB Inode Start</td><td>%d</td></tr>\n", int(TemporalSuperBloque.S_inode_start))
	dot += fmt.Sprintf("<tr><td>SB Block Start</td><td>%d</td></tr>\n", int(TemporalSuperBloque.S_block_start))
	dot += "</table>\n"
	dot += ">];\n"
	dot += "}\n"

	rutaCarpeta := filepath.Dir(path)

	RutaReporte := rutaCarpeta + "/ReporteSuperBloque.dot"
	err = os.WriteFile(RutaReporte, []byte(dot), 0644)
	if err != nil {
		fmt.Println("Error REP DISK: Error al escribir el archivo DOT.")
		return err
	}
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("Error REP DISK: Error al crear el directorio.")
			return err
		}
	}
	cmd := exec.Command("dot", "-Tjpg", RutaReporte, "-o", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error REP DISK: Error al ejecutar Graphviz.")
		return err
	}
	return nil
}

func RepInode(id string, path string) error {
	var ParticionesMontadas DiskManagement.MountedPartition
	var ParticionEncontrada bool

	for _, Particiones := range DiskManagement.GetMountedPartitions() {
		for _, Particion := range Particiones {
			fmt.Println(Particion.ID)
			fmt.Println(id)
			if Particion.ID == id {
				ParticionesMontadas = Particion
				ParticionEncontrada = true
				break
			}
		}
		if ParticionEncontrada {
			break
		}
	}

	if !ParticionEncontrada {
		fmt.Printf("Error REP INODE: No se encontró la partición con el ID: %s.\n", id)
		return fmt.Errorf("no se encontró la partición con el ID: %s", id)
	}

	archivo, err := Utilities.OpenFile(ParticionesMontadas.Path)
	if err != nil {
		fmt.Println("Error REP INODE: Error al abrir el archivo.")
		return err
	}
	defer func(archivo *os.File) {
		err := archivo.Close()
		if err != nil {
			fmt.Println("Error REP INODE: Error al cerrar el archivo.")
		}
	}(archivo)

	var MBRTemporal Structs.MRB
	if err := Utilities.ReadObject(archivo, &MBRTemporal, 0); err != nil {
		fmt.Println("Error REP INODE: Error al leer el MBR.")
		return err
	}

	var index = -1
	for i := 0; i < 4; i++ {
		if MBRTemporal.Partitions[i].Size != 0 {
			if strings.Contains(string(MBRTemporal.Partitions[i].Id[:]), id) {
				if MBRTemporal.Partitions[i].Status[0] == '1' {
					index = i
				} else {
					fmt.Println("Error REP INODE: La partición con el ID:%s no está montada.\n" + id)
					return errors.New("La partición con el ID:%s no está montada.\n" + id)
				}
				break
			}
		}
	}

	if index == -1 {
		fmt.Printf("Error REP INODE: No se encontró la partición con el ID: %s.\n", id)
		return fmt.Errorf("No se encontró la partición con el ID: %s.", id)
	}

	var TemporalSuperBloque = Structs.Superblock{}
	if err := Utilities.ReadObject(archivo, &TemporalSuperBloque, int64(MBRTemporal.Partitions[index].Start)); err != nil {
		fmt.Println("Error REP INODE: Error al leer el SuperBloque.")
		return err
	}

	var dotFile bytes.Buffer

	fmt.Fprintln(&dotFile, "digraph G {")
	fmt.Fprintln(&dotFile, "node [shape=none];")
	fmt.Fprintln(&dotFile, "fontname=\"Courier New\";")
	fmt.Fprintln(&dotFile, "title [label=\"Reporte Inode\"];")

	for i := 0; i < int(TemporalSuperBloque.S_inodes_count); i++ {
		var inode Structs.Inode

		if err := Utilities.ReadObject(archivo, &inode, int64(TemporalSuperBloque.S_inode_start)+int64(i)*int64(TemporalSuperBloque.S_inode_size)); err != nil {
			fmt.Println("Error al leer el inodo:" + string(err.Error()))
			continue
		}

		if inode.I_size > 0 {
			fmt.Fprintf(&dotFile, "inode%d [label=<\n", i)
			fmt.Fprintf(&dotFile, "<table border='0' cellborder='1' cellspacing='0' cellpadding='10'>\n")
			fmt.Fprintf(&dotFile, "<tr><td colspan='2' bgcolor='skyblue'>Inode %d</td></tr>\n", i)
			fmt.Fprintf(&dotFile, "<tr><td>UID</td><td>%d</td></tr>\n", inode.I_uid)
			fmt.Fprintf(&dotFile, "<tr><td>GID</td><td>%d</td></tr>\n", inode.I_gid)
			fmt.Fprintf(&dotFile, "<tr><td>Size</td><td>%d</td></tr>\n", inode.I_size)
			fmt.Fprintf(&dotFile, "<tr><td>ATime</td><td>%s</td></tr>\n", html.EscapeString(string(inode.I_atime[:])))
			fmt.Fprintf(&dotFile, "<tr><td>CTime</td><td>%s</td></tr>\n", html.EscapeString(string(inode.I_ctime[:])))
			fmt.Fprintf(&dotFile, "<tr><td>MTime</td><td>%s</td></tr>\n", html.EscapeString(string(inode.I_mtime[:])))
			fmt.Fprintf(&dotFile, "<tr><td>Blocks</td><td>%v</td></tr>\n", inode.I_block)
			//fmt.Fprintf(&dotFile, "<tr><td>Type</td><td>%c</td></tr>\n", inode.IN_Type[0])
			fmt.Fprintf(&dotFile, "<tr><td>Perms</td><td>%v</td></tr>\n", inode.I_perm)
			fmt.Fprintf(&dotFile, "</table>\n")
			fmt.Fprintf(&dotFile, " >];\n")
		}
	}
	fmt.Fprintln(&dotFile, "}")

	rutaCarpeta := filepath.Dir(path)

	RutaReporte := rutaCarpeta + "/ReporteInode.dot"
	err = os.WriteFile(RutaReporte, dotFile.Bytes(), 0644)
	if err != nil {
		fmt.Println("Error REP DISK: Error al escribir el archivo DOT.")
		return err
	}

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("Error REP INODE: Error al crear el directorio.")
			return err
		}
	}
	cmd := exec.Command("dot", "-Tjpg", RutaReporte, "-o", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error REP INODE: Error al ejecutar Graphviz.")
		return err
	}
	return nil
}

func RepBMInode(id string, path string) error {

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("Error al crear el directorio:" + string(err.Error()))
			return err
		}
	}

	var ParticionesMontadas DiskManagement.MountedPartition
	var ParticionEncontrada bool

	for _, Particiones := range DiskManagement.GetMountedPartitions() {
		for _, Particion := range Particiones {
			if Particion.ID == id {
				ParticionesMontadas = Particion
				ParticionEncontrada = true
				break
			}
		}
		if ParticionEncontrada {
			break
		}
	}

	if !ParticionEncontrada {
		fmt.Printf("Error REP BMInode: No se encontró la partición con el ID: %s.\n", id)
		return fmt.Errorf("No se encontró la partición con el ID: %s.", id)
	}

	archivo, err := Utilities.OpenFile(ParticionesMontadas.Path)
	if err != nil {
		fmt.Println("Error al abrir el archivo:" + string(err.Error()))
		return err
	}
	defer func(archivo *os.File) {
		err := archivo.Close()
		if err != nil {
			fmt.Println("Error al cerrar el archivo:" + string(err.Error()))
		}
	}(archivo)

	var MBRTemporal Structs.MRB
	if err := Utilities.ReadObject(archivo, &MBRTemporal, 0); err != nil {
		fmt.Println("Error al leer el MBR:" + string(err.Error()))
		return err
	}

	var index = -1
	for i := 0; i < 4; i++ {
		if MBRTemporal.Partitions[i].Size != 0 {
			if strings.Contains(string(MBRTemporal.Partitions[i].Id[:]), id) {
				if MBRTemporal.Partitions[i].Status[0] == '1' {
					index = i
				} else {
					fmt.Println("Error REP SB: La partición con el ID:%s no está montada.\n" + id)
					return errors.New("La partición con el ID:%s no está montada.\n" + id)
				}
				break
			}
		}
	}

	if index == -1 {
		fmt.Printf("Error REP SB: No se encontró la partición con el ID: %s.\n", id)
		return fmt.Errorf("No se encontró la partición con el ID: %s.", id)
	}

	var TemporalSuperBloque = Structs.Superblock{}
	if err := Utilities.ReadObject(archivo, &TemporalSuperBloque, int64(MBRTemporal.Partitions[index].Start)); err != nil {
		fmt.Println("Error REP SB: Error al leer el SuperBloque.")
		return err
	}

	bitmapInode := make([]byte, TemporalSuperBloque.S_free_inodes_count)
	if _, err := archivo.ReadAt(bitmapInode, int64(TemporalSuperBloque.S_bm_inode_start)); err != nil {
		fmt.Println("Error: No se pudo leer el bitmap de inodos:" + string(err.Error()))
		return err
	}

	outputFile, err := os.Create(path)
	if err != nil {
		fmt.Println("Error al crear el archivo de reporte:" + string(err.Error()))
		return err
	}
	defer func(outputFile *os.File) {
		err := outputFile.Close()
		if err != nil {
			fmt.Println("Error al cerrar el archivo de reporte:" + string(err.Error()))
		}
	}(outputFile)

	// Escribir el reporte en el archivo de texto
	fmt.Fprintln(outputFile, "Reporte BitMap Inode")
	fmt.Fprintln(outputFile, "---------------------------------------")

	// Mostrar 20 bits por línea
	for i, bit := range bitmapInode {
		if i > 0 && i%20 == 0 {
			// Nueva línea cada 20 bits
			fmt.Fprintln(outputFile)
		}
		fmt.Fprintf(outputFile, "%d ", bit)
	}
	return nil
}

func RepBMBlock(id string, path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("Error al crear el directorio:" + string(err.Error()))
			return err
		}
	}

	var ParticionesMontadas DiskManagement.MountedPartition
	var ParticionEncontrada bool

	for _, Particiones := range DiskManagement.GetMountedPartitions() {
		for _, Particion := range Particiones {
			if Particion.ID == id {
				ParticionesMontadas = Particion
				ParticionEncontrada = true
				break
			}
		}
		if ParticionEncontrada {
			break
		}
	}

	if !ParticionEncontrada {
		fmt.Printf("Error REP BMBlock: No se encontró la partición con el ID: %s.\n", id)
		return fmt.Errorf("No se encontró la partición con el ID: %s.", id)
	}

	archivo, err := Utilities.OpenFile(ParticionesMontadas.Path)
	if err != nil {
		fmt.Print("Error REP BM BLOCK: Error al abrir el archivo.\n")
		return err
	}
	defer func(archivo *os.File) {
		err := archivo.Close()
		if err != nil {
			fmt.Print("Error REP BM BLOCK: Error al cerrar el archivo.\n")
		}
	}(archivo)

	var MBRTemporal Structs.MRB
	if err := Utilities.ReadObject(archivo, &MBRTemporal, 0); err != nil {
		return err
	}

	var index = -1
	for i := 0; i < 4; i++ {
		if MBRTemporal.Partitions[i].Size != 0 {
			if strings.Contains(string(MBRTemporal.Partitions[i].Id[:]), id) {
				if MBRTemporal.Partitions[i].Status[0] == '1' {
					index = i
				} else {
					fmt.Print("Error REP SB: La partición con el ID:%s no está montada.\n" + id)
					return errors.New("La partición con el ID:%s no está montada.\n" + id)
				}
				break
			}
		}
	}

	if index == -1 {
		fmt.Printf("Error REP BMBlock: No se encontró la partición con el ID: %s.\n", id)
		return fmt.Errorf("No se encontró la partición con el ID: %s.", id)
	}

	var TemporalSuperBloque = Structs.Superblock{}
	if err := Utilities.ReadObject(archivo, &TemporalSuperBloque, int64(MBRTemporal.Partitions[index].Start)); err != nil {
		fmt.Println("Error REP SB: Error al leer el SuperBloque.")
		return err
	}

	// Leer el bitmap de bloques desde el archivo binario
	bitmapBlock := make([]byte, TemporalSuperBloque.S_blocks_count)
	if _, err := archivo.ReadAt(bitmapBlock, int64(TemporalSuperBloque.S_bm_block_start)); err != nil {
		fmt.Println("Error: No se pudo leer el bitmap de bloques:" + string(err.Error()))
		return err
	}

	// Crear el archivo de salida para el reporte
	outputFile, err := os.Create(path)
	if err != nil {
		fmt.Println("Error al crear el archivo de reporte:" + string(err.Error()))
		return err
	}
	defer func(outputFile *os.File) {
		err := outputFile.Close()
		if err != nil {
			fmt.Println("Error al cerrar el archivo de reporte:" + string(err.Error()))
		}
	}(outputFile)

	// Escribir el reporte en el archivo de texto
	fmt.Fprintln(outputFile, "Reporte Bitmap de Bloques")
	fmt.Fprintln(outputFile, "-------------------------")

	// Mostrar 20 bits por línea
	for i, bit := range bitmapBlock {
		if i > 0 && i%20 == 0 {
			// Nueva línea cada 20 bits
			fmt.Fprintln(outputFile)
		}
		fmt.Fprintf(outputFile, "%d ", bit)
	}
	return nil
}
