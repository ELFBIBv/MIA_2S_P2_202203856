
import React from "react";

const FileSystemButtons = ({getDiskName,selectedPartition, handleSearch,fetchFiles,results, path, setPath }) => {
    return (

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
                <div className="row">
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

    );
};

export default FileSystemButtons;