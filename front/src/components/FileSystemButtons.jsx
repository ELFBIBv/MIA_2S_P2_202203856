import React, { useRef, useState, useEffect } from "react";
import Editor from "@monaco-editor/react";

const FileSystemButtons = ({ getDiskName, selectedPartition, handleSearch, fetchFile, results, path, setPath }) => {
    const editorRef = useRef(null);
    const [text, setText] = useState("");

    function handleEditorDidMount(editor) {
        editorRef.current = editor;
    }

    function AddText(texto) {
        setText(texto); // Actualiza el estado de text con el contenido de result.Content
    }

    // Efecto para actualizar el editor cuando `text` cambia
    useEffect(() => {
        if (editorRef.current && text !== "") {
            editorRef.current.setValue(text); // Actualiza el editor con el contenido de `text`
        }
    }, [text]);

    return (
        <div>
            {
                text === "" ? (
                    <div className="flex-grow flex flex-col items-center justify-center p-16">
                        <div className="w-full max-w-3xl p-8 bg-white rounded-lg shadow-md">
                            <h2 className="text-2xl font-bold mb-4 text-gray-800">
                                Sistema de Archivos de la Partición {selectedPartition}
                            </h2>
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
                            <div className="row">
                                {results.length > 0 ? (
                                    results.map((result, index) => (
                                        <div key={index} className="col-auto">
                                            {result.Type === "folder" ? (
                                                <div
                                                    className="custom-card"
                                                    style={{ cursor: "pointer", padding: "10px" }}
                                                    onClick={() => fetchFile(result.Name)}
                                                >
                                                    <div id="FolderImage" className="p-1"></div>
                                                    <div>
                                                        <h5 className="card-title text-center" style={{ whiteSpace: "nowrap" }}>
                                                            {getDiskName(result.Name)}
                                                        </h5>
                                                    </div>
                                                </div>
                                            ) : (
                                                <div
                                                    className="custom-card"
                                                    style={{ cursor: "pointer", padding: "10px" }}
                                                    onClick={() => AddText(result.Content)} // Llama a AddText con el contenido
                                                >
                                                    <div id="FileImage" className="p-1"></div>
                                                    <div>
                                                        <h5 className="card-title text-center" style={{ whiteSpace: "nowrap" }}>
                                                            {getDiskName(result.Name)}
                                                        </h5>
                                                    </div>
                                                </div>
                                            )}
                                        </div>
                                    ))
                                ) : (
                                    <p>No se han creado discos aún.</p>
                                )}
                            </div>
                        </div>
                    </div>
                ) : (
                    <div className="input-area">
                        <label htmlFor="code">Entrada:</label>
                        <Editor
                            height="50vh"
                            theme="vs-dark"
                            onMount={handleEditorDidMount}
                            value={text} // El editor muestra el valor de `text`
                        />
                    </div>
                )
            }
        </div>
    );
};

export default FileSystemButtons;
