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

    const runCode = () => {
        cleanOutput();
        const code = editorRef.current.getValue();
        fetch('http://localhost:8080/execute', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ code: code })
        })
            .then((resp) => resp.json())
            .then(data => {
                setOutput(data.output);
                consolaRef.current.setValue(data.output);
            })
            .catch(error => console.error('hubo un problema con la ejecucion del fetch:', error));
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
        const resizeObserver = new ResizeObserver(debounce(entries => {
            // No se necesita lógica específica aquí
        }, 100)); // Ajusta el tiempo según sea necesario

        resizeObserver.observe(document.querySelector('#output'));

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