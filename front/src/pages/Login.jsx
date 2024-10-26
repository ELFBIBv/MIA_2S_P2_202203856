import React, { useState } from "react";
import CommandService from "../services/CommandService";
import Swal from "sweetalert2";

const Login = ({ onLogin }) => {
    const [userId, setUserId] = useState("");
    const [password, setPassword] = useState("");
    const [partitionId, setPartitionId] = useState("");
    const [output, setOutput] = useState("");

    const handleLogin = async () => {
        const loginCommand = `login -user="${userId}" -pass="${password}" -id="${partitionId}"`;

        // CommandService.parseAndSend(loginCommand)
        //     .then((response) => {
        //         console.log("Respuesta completa del servidor:", response);
        //
        //         // Guardar el usuario en localStorage
        //         localStorage.setItem("loggedUser", userId);
        //         setOutput(`Login exitoso: ${JSON.stringify(response)}`);
        //         Swal.fire("Inicio de sesión exitoso", "Bienvenido a la aplicación", "success");
        //
        //         // Llamar a la función pasada como prop para notificar el login
        //         if (onLogin) onLogin(userId);
        //     })
        //     .catch((error) => {
        //         console.error("Error completo recibido:", error);
        //         const errorMessage = error.response?.data?.error || "Error desconocido";
        //         Swal.fire("Error al iniciar sesión", errorMessage, "error");
        //         setOutput(`Error: ${errorMessage}`);
        //     });
        try {
            const response = await fetch('http://localhost:8080/execute', {
                method: 'POST',
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({code: loginCommand})
            });

            // Si el estado no es ok, arroja un error con el contenido de la respuesta
            if (!response.ok) {
                const data = await response.json();
                setOutput(data.output);
                await Swal.fire("Error al iniciar sesión", data.output, "error");
                return;
            }

            const data = await response.json();

            if (data.output.split(":")[0] === "Error") {
                setOutput(data.output);
                await Swal.fire("Error al iniciar sesión", data.output, "error");
                return;
            }

            // Devolver la respuesta en JSON si todo está bien
            await Swal.fire("Inicio de sesión exitoso", "Bienvenido a la aplicación", "success");
            setOutput(data.output);
        } catch (error) {
            // Capturar cualquier error durante la solicitud
            throw new Error(error.message || "Error al enviar el comando");
        }
    };

    return (
        <div className="container mt-5">
            <h2>Login</h2>
            <div className="mb-3">
                <label>ID Partición</label>
                <input
                    type="text"
                    className="form-control"
                    value={partitionId}
                    onChange={(e) => setPartitionId(e.target.value)}
                    placeholder="Ingrese el ID de la partición"
                />
            </div>
            <div className="mb-3">
                <label>Usuario</label>
                <input
                    type="text"
                    className="form-control"
                    value={userId}
                    onChange={(e) => setUserId(e.target.value)}
                    placeholder="Ingrese su usuario"
                />
            </div>
            <div className="mb-3">
                <label>Contraseña</label>
                <input
                    type="password"
                    className="form-control"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Ingrese su contraseña"
                />
            </div>
            <button className="btn btn-primary" onClick={handleLogin}>
                Iniciar sesión
            </button>
            <pre>{output}</pre>
        </div>
    );
};

export default Login;