import React, { useState, useEffect } from "react";
import './Visualizador.css';
import DisksButtons from "../components/DisksButtons";
import PartitionsButtons from "../components/PartitionsButtons";
import Swal from "sweetalert2";
import FileSystemButtons from "../components/FileSystemButtons";

const Visualizador = () => {
    const [disks, setDisks] = useState([]);
    const [selectedDisk, setSelectedDisk] = useState("");
    const [partitions, setPartitions] = useState([]);
    const [selectedPartition, setSelectedPartition] = useState("");
    const [path, setPath] = useState("/");
    const [results, setResults] = useState([]);
    /*
    * la estructura de los result va a tener el tipo
    * {
    *  name: "nombre",
    * type: "folder" | "file"
    * }
    * */

    useEffect(() => {
        // Hacer la solicitud al backend
        fetch('http://localhost:8080/getPathDisks')
            .then(response => response.json())
            .then(data => {
                // Guardar el array de strings en el estado
                const storedDisks = data || [];
                setDisks(storedDisks);
            })
            .catch(error => console.error('Error:', error));
    }, []);

    // Función para obtener solo el nombre del archivo del path
    const getDiskName = (path) => {
        if (typeof path === 'string') {
            return path.split('/').pop();
        } else {
            // Si no es una cadena, devuelve un valor por defecto o maneja el error
            console.error("path no es una cadena:", path);
            return "error";
        }
        // return path.split('/').pop();
    };

    // Función para obtener las particiones de un disco
    const fetchPartitions = (diskPath) => {
        // Guardar el disco seleccionado
        setSelectedDisk(diskPath);

        // Hacer la solicitud al backend para obtener las particiones
        fetch("http://localhost:8080/getMountedPartitionsForPathDisk", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({path: diskPath}),
        })
            .then((response) => response.json())
            .then((data) => {
                setPartitions(data || []);
            })
            .catch((error) => {
                console.error("Error al obtener particiones:", error);
                setPartitions([]);
            });
    };

    const fetchFilesSystem = (selectedPartition) => {
        // Guardar la partición seleccionada
        setPath("/");
        setSelectedPartition(selectedPartition);
        console.log("FilePath: ", path);
        console.log("DiskPath: ", selectedDisk);
        console.log("PartitionName: ", selectedPartition);

        // Hacer la solicitud al backend para obtener los archivos
        fetch("http://localhost:8080/readfiles", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({FilePath: path,DiskPath: selectedDisk, PartitionName: selectedPartition}),
        })
            .then((response) => response.json())
            .then((data) => {
                setResults(data || []);
            })
            .catch((error) => {
                console.error("Error al obtener archivos:", error);
                setResults([]);
            });
    };

    const fetchFiles = (fileSelected) => {
        // Guardar la partición seleccionada
        setPath(path + fileSelected);

        // Hacer la solicitud al backend para obtener los archivos
        fetch("http://localhost:8080/readfiles", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({FilePath: path,DiskPath: selectedDisk, PartitionName: selectedPartition}),
        })
            .then((response) => response.json())
            .then(async (data) => {
                // await Swal.fire("Error al iniciar sesión", data, "error");
                setResults(data || []);
            })
            .catch((error) => {
                console.error("Error al obtener archivos:", error);
                setResults([]);
            });
    };

    const handleSearch = () => {
        // Simular resultados de búsqueda
        // vamos a mandar una solicitud al backend para obtener los archivos de la ruta
    };

    return (
        <div className="container mt-5">

            {!selectedDisk && (
                <div>
                    <h2>Discos Creados</h2>
                <DisksButtons disks={disks} fetchPartitions={fetchPartitions} getDiskName={getDiskName}/>
                </div>
            )}


            {/* Mostrar las particiones del disco seleccionado */}
            {selectedDisk && !selectedPartition && (
                <div>
                    <h2>Particiones del Disco: {getDiskName(selectedDisk)}</h2>
                    <PartitionsButtons partitions={partitions} fetchFilesSystem={fetchFilesSystem}/>
                </div>
            )}
            {/* Mostrar los archivos de la partición seleccionada */}
            {selectedPartition && (
                // {getDiskName,selectedPartition, handleSearch,fetchFiles,results }
                <FileSystemButtons getDiskName={getDiskName}
                                   selectedPartition={selectedPartition}
                                   handleSearch={handleSearch}
                                   fetchFiles={fetchFiles}
                                   results={results}
                                   path={path}
                                   setPath={setPath}
                />
            )}

        </div>
    );
};

export default Visualizador;