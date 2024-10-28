import React, { useState, useEffect } from "react";
import './Visualizador.css';
import DisksButtons from "../components/DisksButtons";
import PartitionsButtons from "../components/PartitionsButtons";
import Swal from "sweetalert2";

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
        fetch("http://localhost:8080/readmbr", {
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
            {selectedDisk && (
                <div>
                    <h2>Particiones del Disco: {getDiskName(selectedDisk)}</h2>
                    <PartitionsButtons partitions={partitions} fetchFilesSystem={fetchFilesSystem}/>
                </div>
            )}
            {/* Mostrar los archivos de la partición seleccionada */}
            {selectedPartition && (
                <div className="flex-grow flex flex-col items-center justify-center p-16">
                    <div className="w-full max-w-3xl p-8 bg-white rounded-lg shadow-md">
                        <h2 className="text-2xl font-bold mb-4 text-gray-800">
                            Sistema de Archivos de la Partición {selectedPartition}
                        </h2>
                        {/* Información de la partición */}
                        <p className="text-gray-700 mb-4">Sistema de Archivos: {"NTFS"}</p>
                        <div className="flex mb-4">
                            <input
                                type="text"
                                value={path}
                                onChange={(e) => setPath(e.target.value)}
                                placeholder="Ingrese el path"
                                className="flex-grow p-2 border border-gray-300 rounded-l-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                            <button
                                onClick={handleSearch}
                                className="p-2 bg-blue-500 text-white rounded-r-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500"
                            >
                                Buscar
                            </button>
                        </div>
                        {/* Resultados de la búsqueda */}
                        <div className="flex flex-wrap gap-4">
                            {results.length > 0 ? (
                                results.map((result, index) => (
                                    <div key={index} className="col-auto"> {/* col-auto ajusta el tamaño automático */}
                                        {/* Mostrar el nombre del archivo */}
                                        <div
                                            className="custom-card" /* Clase personalizada */
                                            style={{cursor: "pointer", padding: "10px"}} /* Reducir padding */
                                            onClick={() => fetchFiles(result.Name)}
                                        >
                                            <div id="DiskImage" className="p-1"> {/* Ajustar padding de la imagen */}
                                            </div>
                                            <div>
                                                <h5 className="card-title text-center" style={{whiteSpace: "nowrap"}}>
                                                    {getDiskName(result.Name)}
                                                </h5>
                                            </div>
                                        </div>
                                    </div>
                                ))
                            ) : (
                                <p>No se han creado discos aún.</p>
                            )}
                        </div>
                    </div>
                </div>
            )}

        </div>
    );
};

export default Visualizador;