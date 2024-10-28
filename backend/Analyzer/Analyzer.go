package Analyzer

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"proyecto1/AdminFiles"
	"proyecto1/Commands"
	"proyecto1/User"
	"regexp"
	"strings"
)

var Salida = ""
var LoggedPartitionID = ""
var UserPermissions = [3]byte{'0', '0', '0'}

// inicializar un array de strings para guardar los paths de los discos
var PathDisksMounted []string

func getCommandAndParams(input string) (string, string, string) {
	reIgnorarComentarios := regexp.MustCompile(`#.*`)
	comentario := ""
	comentario = reIgnorarComentarios.FindString(input)
	if comentario != "" {
		comentario = "  " + comentario
	}
	input = reIgnorarComentarios.ReplaceAllString(input, "")
	parts := strings.Fields(input)
	if len(parts) > 0 {
		command := strings.ToLower(parts[0])
		params := strings.Join(parts[1:], " ")
		return command, params, comentario
	}
	return "", input, comentario
}

func Analyze(Script string) {
	// leemos linea por linea del script
	Scripts := strings.Split(Script, "\n")
	Salida = ""
	PathDisksMounted = []string{}

	re := regexp.MustCompile(`-(\w+)(="[^"]+"|=\w+|=-\w+|=/\S+)?`)

	for _, v := range Scripts {
		//var input string
		if strings.TrimSpace(v) == "" {
			Salida += "\n"
			continue
		}

		command, params, comentario := getCommandAndParams(v)
		if command == "" {
			Salida += "\n" + comentario
			continue
		}

		fmt.Println("======================")
		fmt.Println("comando ingresado: ", command, params)

		matches := re.FindAllStringSubmatch(params, -1)
		var err error = nil
		if strings.Contains(command, "mkdisk") {
			err = fnMkdisk(matches)
		} else if strings.Contains(command, "rmdisk") {
			err = fnRmdisk(matches)
		} else if strings.Contains(command, "fdisk") {
			err = fnFdisk(matches)
		} else if strings.Contains(command, "mount") {
			err = fnMount(matches)
		} else if strings.Contains(command, "mkfs") {
			err = fnMkfs(matches)
		} else if strings.Contains(command, "cat") {
			err = fnCat(matches)
		} else if strings.Contains(command, "login") {
			err = fnLogin(matches)
		} else if strings.Contains(command, "logout") {
			err = fnLogout(matches)
		} else if strings.Contains(command, "mkgrp") {
			err = fnMkgrp(matches)
		} else if strings.Contains(command, "rmgrp") {
			err = fnRmgrp(matches)
		} else if strings.Contains(command, "mkusr") {
			err = fnMkusr(matches)
		} else if strings.Contains(command, "rmusr") {
			err = fnRmusr(matches)
		} else if strings.Contains(command, "chgrp") { //este comando no esta implementado
			fmt.Println("chgrp no implementado")
			err = errors.New("chgrp no implementado")
		} else if strings.Contains(command, "mkfile") {
			err = fnMkfile(matches)
		} else if strings.Contains(command, "mkdir") {
			err = fnMkdir(matches)
		} else if strings.Contains(command, "rep") {
			err = fnRep(matches)
		} else {
			err = errors.New("Commando '" + command + "' invalido o no encontrado")
		}
		if err != nil {
			Salida += fmt.Sprintf("\nError: %s", err)
		}
		if comentario != "" {
			Salida += comentario
		}
	}
	Salida = strings.TrimSpace(Salida)

	fmt.Println("fin de la ejecución")
}

func fnMkdisk(matches [][]string) error {
	// Definir flag
	fs := flag.NewFlagSet("mkdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	fit := fs.String("fit", "ff", "Ajuste")
	unit := fs.String("unit", "m", "Unidad")
	path := fs.String("path", "", "Ruta")

	// Parse flag (quitamos el primer valor de las flags el cual es el path del archivo)
	err := fs.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	// Encontrar la flag en el input

	// Process the input
	for _, match := range matches {
		fmt.Println("match: ", match)

		flagName := strings.ToLower(match[1])  // match[1]: Captura y guarda el nombre del flag (por ejemplo, "size", "unit", "fit", "path")
		flagValue := strings.ToLower(match[2]) //trings.ToLower(match[2]): Captura y guarda el valor del flag, asegurándose de que esté en minúsculas
		if flagName == "path" {
			flagValue = match[2]
		}
		flagValue = strings.Trim(flagValue, "=")
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "size", "fit", "unit", "path":
			err := fs.Set(flagName, flagValue)
			if err != nil {
				return err
			}
		default:
			return errors.New("Flag '" + flagName + "' not found")
		}
	}

	// LLamamos a la funcion
	if err := Commands.Mkdisk(*size, *fit, *unit, *path); err != nil {
		return err
	}
	Salida += fmt.Sprintf("\nDisco creado con éxito en la ruta: %s", *path)
	return nil
}

func fnRmdisk(matches [][]string) error {
	// Definir flag
	fs := flag.NewFlagSet("rmdisk", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")

	// Parse flag (quitamos el primer valor de las flags el cual es el path del archivo)
	err := fs.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	// Encontrar la flag en el input

	// Process the input
	for _, match := range matches {
		flagName := strings.ToLower(match[1])  // match[1]: Captura y guarda el nombre del flag (por ejemplo, "size", "unit", "fit", "path")
		flagValue := strings.ToLower(match[2]) //trings.ToLower(match[2]): Captura y guarda el valor del flag, asegurándose de que esté en minúsculas
		if flagName == "path" {
			flagValue = match[2]
		}

		flagValue = strings.Trim(flagValue, "=")
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "path":
			err := fs.Set(flagName, flagValue)
			if err != nil {
				return err
			}
		default:
			return errors.New("Flag '" + flagName + "' not found")
		}
	}

	// LLamamos a la funcion
	if err := Commands.Rmdisk(*path); err != nil {
		return err
	}
	Salida += fmt.Sprintf("\nDisco eliminado con éxito en la ruta: %s", *path)
	return nil
}

func fnFdisk(matches [][]string) error {
	// Definir flags
	fs := flag.NewFlagSet("fdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	unit := fs.String("unit", "k", "Unidad")
	path := fs.String("path", "", "Ruta")
	type_ := fs.String("type", "p", "Tipo")
	fit := fs.String("fit", "wf", "Ajuste")
	name := fs.String("name", "", "Nombre")
	delete_ := fs.String("delete", "", "Eliminar partición (Fast/Full)")
	add := fs.Int("add", 0, "Agregar espacio a la partición")
	primero := ""

	// Parsear los flags
	err := fs.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	// Procesar el input
	for _, match := range matches {
		flagName := strings.ToLower(match[1])
		flagValue := strings.ToLower(match[2]) //trings.ToLower(match[2]): Captura y guarda el valor del flag, asegurándose de que esté en minúsculas
		if flagName == "path" {
			flagValue = match[2]
		}

		flagValue = strings.Trim(flagValue, "=")
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "size", "fit", "unit", "path", "name", "type", "delete", "add":
			switch flagName {
			case "size", "delete", "add":
				if primero != "" {
					continue
				}
				primero = flagName
			}
			err := fs.Set(flagName, flagValue)
			if err != nil {
				return err
			}
		default:
			return errors.New("Flag '" + flagName + "' not found")
		}
	}

	// Llamar a la función
	if err := Commands.Fdisk(*size, *path, *name, *unit, *type_, *fit, *delete_, *add); err != nil {
		return err
	}
	Salida += fmt.Sprintf("\nPartición creada con éxito en la ruta: %s", *path)
	return nil
}

func fnMount(matches [][]string) error {
	fs := flag.NewFlagSet("mount", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")
	name := fs.String("name", "", "Nombre de la partición")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	for _, match := range matches {
		flagName := strings.ToLower(match[1])
		flagValue := strings.ToLower(match[2]) //trings.ToLower(match[2]): Captura y guarda el valor del flag, asegurándose de que esté en minúsculas
		if flagName == "path" {
			flagValue = match[2]
		}

		flagValue = strings.Trim(flagValue, "=")
		flagValue = strings.Trim(flagValue, "\"") //quitamos todos los " del string

		switch flagName {
		case "path", "name":
			err := fs.Set(flagName, flagValue)
			if err != nil {
				return err
			}
		default:
			return errors.New("Flag '" + flagName + "' not found")
		}
	}

	//ejecutamos el comando
	if err := Commands.Mount(*path, *name); err != nil {
		return err
	}
	Salida += fmt.Sprintf("\nPartición montada con éxito en la ruta: %s", *path)
	//agregamos las particiones montadas
	PathDisksMounted = append(PathDisksMounted, *path)
	fmt.Println("PathDisksMounted: ", PathDisksMounted)
	return nil
}

func fnMkfs(matches [][]string) error {
	fs := flag.NewFlagSet("mkfs", flag.ExitOnError)
	id := fs.String("id", "", "Id")
	type_ := fs.String("type", "full", "Tipo")

	// Procesar los matches
	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		flagValue = strings.ToLower(flagValue)
		flagName = strings.ToLower(flagName)

		flagValue = strings.Trim(flagValue, "=")
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "id", "type":
			err := fs.Set(flagName, flagValue)
			if err != nil {
				return err
			}
		default:
			fmt.Println("Error: Flag not found")
			return errors.New("Flag '" + flagName + "' not found")
		}
	}

	// Verifica que se hayan establecido todas las flags necesarias
	if *id == "" {
		return errors.New("id es un parámetro obligatorio")
	}

	if *type_ == "" {
		err := fs.Set("type", "full")
		if err != nil {
			return err
		}
	}

	// Llamar a la función
	if err := Commands.Mkfs(*id, *type_); err != nil {
		return err
	}
	Salida += fmt.Sprintf("\nSistema de archivos creado con éxito en la partición con id: %s", *id)
	return nil
}

func fnCat(matches [][]string) error {

	var files []string
	// Procesar el input
	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		flagValue = strings.ToLower(flagValue)
		flagName = strings.ToLower(flagName)

		flagValue = strings.Trim(flagValue, "=")
		flagValue = strings.Trim(flagValue, "\"")

		fmt.Println("flagName: ", flagName)
		fmt.Println("flagValue: ", flagValue)

		expr := regexp.MustCompile(`^file\d*$`)

		name := expr.FindAllStringSubmatch(flagName, -1)

		if len(name) == 0 {
			return errors.New("Flag '" + flagName + "' not found")
		}

		files = append(files, flagValue)
	}

	if len(files) == 0 {
		return errors.New("no se especificaron archivos")
	}

	if LoggedPartitionID == "" || UserPermissions == [3]byte{'0', '0', '0'} {
		return errors.New("no hay ningún usuario logueado")
	}

	fmt.Println("paths:", files)

	err, text := Commands.Cat(files, LoggedPartitionID, UserPermissions)

	if err != nil {
		return err
	}

	Salida += "\n" + text
	return nil
}

func fnLogin(matches [][]string) error {
	// Definir las flags
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	user := fs.String("user", "", "Usuario")
	pass := fs.String("pass", "", "Contraseña")
	id := fs.String("id", "", "Id")

	// Parsearlas
	err := fs.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	// Procesar el input
	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		flagValue = strings.Trim(flagValue, "=")
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "user", "pass", "id":
			if flagName == "id" {
				flagValue = strings.ToLower(flagValue)
			}
			err := fs.Set(flagName, flagValue)
			if err != nil {
				return err
			}
		default:
			fmt.Println("Error: Flag not found")
			return errors.New("Flag '" + flagName + "' not found")
		}
	}

	// Llamar a la función
	err, permisos := User.Login(*user, *pass, *id)
	if err != nil {
		return err
	}
	Salida += fmt.Sprintf("\nUsuario logueado con éxito con el id: %s", *id)
	LoggedPartitionID = *id
	UserPermissions = permisos

	return nil
}

func fnLogout(matches [][]string) error {
	if len(matches) > 0 {
		return errors.New("no se permiten parámetros para el comando logout")
	}
	//como pongo la barra vertical en el teclado?
	//pon las barras verticales expresando el or para ver si tambien no hay permisos
	if LoggedPartitionID == "" || UserPermissions == [3]byte{'0', '0', '0'} {
		return errors.New("no hay ningún usuario logueado")
	}

	if err := User.Logout(LoggedPartitionID); err != nil {
		return err
	}

	LoggedPartitionID = ""
	UserPermissions = [3]byte{'0', '0', '0'}

	Salida += fmt.Sprintf("\nUsuario deslogueado con éxito con el id: %s", LoggedPartitionID)
	return nil
}

func fnMkgrp(matches [][]string) error {
	fs := flag.NewFlagSet("mkgrp", flag.ExitOnError)
	name := fs.String("name", "", "Nombre")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		flagValue = strings.Trim(flagValue, "=")
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "name":
			fmt.Println("name: ", flagValue)
			err := fs.Set(flagName, flagValue)
			if err != nil {
				return err
			}
		default:
			return errors.New("Flag '" + flagName + "' not found")
		}
	}

	if *name == "" {
		return errors.New("el nombre del grupo es un parámetro obligatorio")
	}

	if len(*name) > 10 {
		return errors.New("el nombre del grupo no puede tener más de 10 caracteres")
	}

	if LoggedPartitionID == "" || UserPermissions == [3]byte{'0', '0', '0'} {
		return errors.New("no hay ningún usuario logueado")
	}

	if !(UserPermissions == [3]byte{'7', '7', '7'}) {
		return errors.New("no tienes permisos para ejecutar esta acción")
	}

	err = User.Mkgrp(*name, LoggedPartitionID)
	if err != nil {
		return err
	}
	Salida += fmt.Sprintf("\nGrupo creado con éxito con el nombre: %s", *name)
	return nil
}

func fnRmgrp(matches [][]string) error {
	fs := flag.NewFlagSet("rmgrp", flag.ExitOnError)
	name := fs.String("name", "", "Nombre")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		flagName = strings.ToLower(flagName)
		flagValue = strings.ToLower(flagValue)

		flagValue = strings.Trim(flagValue, "=")
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "name":
			fmt.Println("name: ", flagValue)
			err := fs.Set(flagName, flagValue)
			if err != nil {
				return err
			}
		default:
			return errors.New("Flag '" + flagName + "' not found")
		}
	}

	if *name == "" {
		return errors.New("el nombre del grupo a eliminar es un parámetro obligatorio")
	}

	if len(*name) > 10 {
		return errors.New("el nombre del grupo a eliminar no puede tener más de 10 caracteres")
	}

	if LoggedPartitionID == "" || UserPermissions == [3]byte{'0', '0', '0'} {
		return errors.New("no hay ningún usuario logueado")
	}

	if !(UserPermissions == [3]byte{'7', '7', '7'}) {
		return errors.New("no tienes permisos para ejecutar esta acción")
	}

	err = User.Rmgrp(*name, LoggedPartitionID)
	if err != nil {
		return err
	}
	Salida += fmt.Sprintf("\nGrupo con nombre '%s' eliminado con éxito", *name)
	return nil
}

func fnMkusr(matches [][]string) error {
	fs := flag.NewFlagSet("mkusr", flag.ExitOnError)
	user := fs.String("user", "", "Usuario")
	pass := fs.String("pass", "", "Contraseña")
	grp := fs.String("grp", "", "Grupo")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		//flagName = strings.ToLower(flagName)
		//flagValue = strings.ToLower(flagValue)

		flagValue = strings.Trim(flagValue, "=")
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "user", "pass", "grp":
			err := fs.Set(flagName, flagValue)
			if err != nil {
				return err
			}
		default:
			return errors.New("Flag '" + flagName + "' not found")
		}
	}

	if *user == "" {
		return errors.New("el nombre del usuario es un parámetro obligatorio")
	}

	if len(*user) > 10 {
		return errors.New("el nombre del usuario no puede tener más de 10 caracteres")
	}

	if *pass == "" {
		return errors.New("la contraseña del usuario es un parámetro obligatorio")
	}

	if len(*pass) > 10 {
		return errors.New("la contraseña del usuario no puede tener más de 10 caracteres")
	}

	if *grp == "" {
		return errors.New("el grupo del usuario es un parámetro obligatorio")
	}

	if len(*grp) > 10 {
		return errors.New("el grupo del usuario no puede tener más de 10 caracteres")
	}

	if LoggedPartitionID == "" || UserPermissions == [3]byte{'0', '0', '0'} {
		return errors.New("no hay ningún usuario logueado")
	}

	if !(UserPermissions == [3]byte{'7', '7', '7'}) {
		return errors.New("no tienes permisos para ejecutar esta acción")
	}

	err = User.Mkusr(*user, *pass, *grp, LoggedPartitionID)
	if err != nil {
		return err
	}
	Salida += fmt.Sprintf("\nUsuario con nombre '%s' creado con éxito", *user)
	return nil
}

func fnRmusr(matches [][]string) error {
	fs := flag.NewFlagSet("rmusr", flag.ExitOnError)
	user := fs.String("user", "", "Usuario")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		flagValue = strings.Trim(flagValue, "=")
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "user":
			err := fs.Set(flagName, flagValue)
			if err != nil {
				return err
			}
		default:
			return errors.New("Flag '" + flagName + "' not found")
		}
	}

	if *user == "" {
		return errors.New("el nombre del usuario a eliminar es un parámetro obligatorio")
	}

	if len(*user) > 10 {
		return errors.New("el nombre del usuario a eliminar no puede tener más de 10 caracteres")
	}

	if LoggedPartitionID == "" || UserPermissions == [3]byte{'0', '0', '0'} {
		return errors.New("no hay ningún usuario logueado")
	}

	if !(UserPermissions == [3]byte{'7', '7', '7'}) {
		return errors.New("no tienes permisos para ejecutar esta acción")
	}

	err = User.Rmusr(*user, LoggedPartitionID)
	if err != nil {
		return err
	}
	Salida += fmt.Sprintf("\nUsuario con nombre '%s' eliminado con éxito", *user)
	return nil
}

func fnMkfile(matches [][]string) error {
	fs := flag.NewFlagSet("mkfile", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")
	r := fs.Bool("r", false, "Se crean las carpetas padres si no existen")
	size := fs.Int("size", 0, "Tamaño")
	cont := fs.String("cont", "", "Contenido")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	for _, match := range matches {
		flagName := match[1]
		flagName = strings.ToLower(flagName)

		if match[2] == "" {
			if flagName == "r" {
				err := fs.Set(flagName, "true")
				if err != nil {
					return err
				}
				continue
			}
		}

		flagValue := strings.ToLower(match[2]) //trings.ToLower(match[2]): Captura y guarda el valor del flag, asegurándose de que esté en minúsculas
		if flagName == "path" {
			flagValue = match[2]
		}

		flagValue = strings.Trim(flagValue, "=")
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "path", "size", "cont":
			err := fs.Set(flagName, flagValue)
			if err != nil {
				return err
			}
		default:
			return errors.New("Flag '" + flagName + "' not found")
		}
	}

	//vamos a validar dentro de la funcion
	err = AdminFiles.Mkfile(*path, *r, *size, *cont, LoggedPartitionID)
	if err != nil {
		return err
	}
	Salida += fmt.Sprintf("\nArchivo creado con éxito en la ruta: %s", *path)
	return nil
}

func fnMkdir(matches [][]string) error {
	fs := flag.NewFlagSet("mkdir", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")
	p := fs.Bool("p", false, "Se crean las carpetas padres si no existen")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	for _, match := range matches {
		flagName := match[1]
		flagName = strings.ToLower(flagName)
		if match[2] == "" {
			if flagName == "p" {
				err := fs.Set(flagName, "true")
				if err != nil {
					return err
				}
				continue
			}
		}

		flagValue := strings.ToLower(match[2]) //trings.ToLower(match[2]): Captura y guarda el valor del flag, asegurándose de que esté en minúsculas
		if flagName == "path" {
			flagValue = match[2]
		}

		flagValue = strings.Trim(flagValue, "=")
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "path":
			err := fs.Set(flagName, flagValue)
			if err != nil {
				return err
			}
		default:
			return errors.New("Flag '" + flagName + "' not found")
		}
	}

	//vamos a validar dentro de la funcion
	err = AdminFiles.Mkdir(*path, *p, LoggedPartitionID)
	if err != nil {
		return err
	}
	Salida += fmt.Sprintf("\nDirectorio creado con éxito en la ruta: %s", *path)
	return nil
}

func fnRep(matches [][]string) error {
	fs := flag.NewFlagSet("rep", flag.ExitOnError)
	name := fs.String("name", "", "Nombre")
	path := fs.String("path", "", "Ruta")
	id := fs.String("id", "", "Id")
	path_file_ls := fs.String("path_file_ls", "", "Ruta del archivo a listar")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	for _, match := range matches {
		flagName := match[1]

		flagName = strings.ToLower(flagName)
		flagValue := strings.ToLower(match[2]) //trings.ToLower(match[2]): Captura y guarda el valor del flag, asegurándose de que esté en minúsculas
		if flagName == "path" {
			flagValue = match[2]
		}

		flagValue = strings.Trim(flagValue, "=")
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "name", "path", "id", "path_file_ls":
			if flagName == "id" {
				flagValue = strings.ToLower(flagValue)
			}
			err := fs.Set(flagName, flagValue)
			if err != nil {
				return err
			}
		default:
			return errors.New("Flag '" + flagName + "' not found")
		}
	}

	if *name == "" {
		return errors.New("el nombre del reporte es un parámetro obligatorio")
	}

	if *path == "" {
		return errors.New("la ruta del reporte es un parámetro obligatorio")
	}

	if *id == "" {
		return errors.New("el id de la partición es un parámetro obligatorio")
	}

	if *path_file_ls == "" && (*name == "ls" || *name == "file") {
		return errors.New("la ruta del archivo a listar es un parámetro obligatorio")
	}

	switch *name {
	case "mbr":
		err := Commands.RepMBR(*id, *path)
		if err != nil {
			return err
		}
		Salida += fmt.Sprintf("\nReporte MBR creado con éxito en la ruta: %s", *path)
	case "disk":
		err := Commands.RepDisk(*id, *path)
		if err != nil {
			return err
		}
		Salida += fmt.Sprintf("\nReporte Disk creado con éxito en la ruta: %s", *path)
	case "inode":
		err := Commands.RepInode(*id, *path)
		if err != nil {
			return err
		}
		Salida += fmt.Sprintf("\nReporte Inode creado con éxito en la ruta: %s", *path)
	case "block":
		fmt.Println("Reporte Block no implementado")
		return errors.New("reporte Block no implementado")
	case "bm_inode":
		err := Commands.RepBMInode(*id, *path)
		if err != nil {
			return err
		}
		Salida += fmt.Sprintf("\nReporte BM Inode creado con éxito en la ruta: %s", *path)
	case "bm_block":
		err := Commands.RepBMBlock(*id, *path)
		if err != nil {
			return err
		}
		Salida += fmt.Sprintf("\nReporte BM Block creado con éxito en la ruta: %s", *path)
	case "sb":
		err := Commands.RepSB(*id, *path)
		if err != nil {
			return err
		}
		Salida += fmt.Sprintf("\nReporte SB creado con éxito en la ruta: %s", *path)
	case "file":
		fmt.Println("Reporte File no implementado")
		return errors.New("reporte File no implementado")
	case "ls":
		fmt.Println("Reporte Ls no implementado")
		return errors.New("reporte Ls no implementado")
	default:
		return errors.New("Reporte '" + *name + "' no encontrado")
	}
	return nil
}
