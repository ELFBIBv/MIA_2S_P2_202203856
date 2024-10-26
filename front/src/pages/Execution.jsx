import React, {useEffect, useRef, useState} from "react";
import Editor from "@monaco-editor/react";
import {debounce} from "lodash";
import './Execution.css';

const Execution = () => {

    const editorRef = useRef(null);
    const consolaRef = useRef(null);
    const [output, setOutput] = useState();
    const valorDefault = "# Aca se veran todos los mensajes de la ejecución";
    const [fileName, setFileName] = useState('');

    function handleEditorDidMount(editor) {
        editorRef.current = editor;
    }

    function handleEditorDidMount2(editor) {
        consolaRef.current = editor;
    }

    const runCode = async () => {
        cleanOutput();
        const code = editorRef.current.getValue();
        if (!code) {
            return;
        }
        try {
            const response = await fetch('http://localhost:8080/execute', {
                method: 'POST',
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({ code: code })
            });

            // Si el estado no es ok, arroja un error con el contenido de la respuesta
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || `Error en el servidor: ${response.statusText}`);
            }

            // Devolver la respuesta en JSON si todo está bien
            const data = await response.json();
            setOutput(data.output);
            consolaRef.current.setValue(data.output);
        } catch (error) {
            // Capturar cualquier error durante la solicitud
            throw new Error(error.message || "Error al enviar el comando");
        }
    };

    const cleanOutput = () => {
        setOutput(valorDefault);
        consolaRef.current.setValue("");
    };

    const uploadFile = (e) => {
        const file = e.target.files[0];
        if (!file) return;

        setFileName(file.name);
        const filereader = new FileReader();

        filereader.readAsText(file);

        filereader.onload = () => {
            editorRef.current.setValue(filereader.result);
        }

        filereader.onerror = () => {
            console.log(filereader.error);
        }
    };


    const clean = () => {
        // Limpiar editor
        editorRef.current.setValue("");

        // Limpiar consola
        setOutput(valorDefault); // Resetear la consola a su valor por defecto

        // Limpiar nombre del archivo cargado
        setFileName(''); // Borrar el nombre del archivo
    };


    useEffect(() => {
        // Establecer un umbral para cambios significativos en tamaño
        const resizeObserver = new ResizeObserver(debounce(entries => {
            const { contentRect } = entries[0];
            // Verificar si hay un cambio significativo de tamaño (más de 20px)
            if (Math.abs(contentRect.width - contentRect.height) > 20) {
                console.log("El contenedor ha cambiado de tamaño significativamente.");
                // Aquí podrías colocar lógica adicional si fuera necesario
            }
        }, 500)); // Un debounce de 300 ms suele ser más que suficiente

        const outputElement = document.querySelector('#output');
        if (outputElement) {
            resizeObserver.observe(outputElement);
        }

        return () => {
            resizeObserver.disconnect();
        };
    }, []);

    return (
        <div className="container">
            <div className="header">
                <div className="buttons">
                    <label htmlFor="file-upload" className="custom-file-upload">Elegir archivo</label>
                    <input id="file-upload" type="file" accept=".smia" style={{display: 'none'}} onChange={uploadFile}/>
                    <input type="button" value="Ejecutar" id="btnEjecutar" className="btn" onClick={runCode}/>
                    <input type="button" value="Limpiar" id="btnLimpiar" className="btn" onClick={clean}/>
                </div>
                <span id="file-name">{fileName || 'No se ha seleccionado archivo'}</span>
            </div>


            <div className="input-area">
                <label htmlFor="code">Entrada:</label>
                <Editor height="50vh" theme="vs-dark" onMount={handleEditorDidMount}/>
            </div>

            <div className="output-area" id="output">
                <label htmlFor="output">Salida:</label>
                <Editor height="28vh"
                        defaultLanguage="plaintext"
                        value={output}
                        defaultValue={valorDefault}
                        theme="vs-dark"
                        options={{readOnly: true}}
                        onMount={handleEditorDidMount2}/>
            </div>
        </div>
    );
};

export default Execution;