import React from 'react';
import { Link } from 'react-router-dom';
import Swal from 'sweetalert2';

const Navbar = () => {
    const handleLogout = async () => {
        try {
            const response = await fetch('http://44.200.113.1:8080/execute', {
                method: 'POST',
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({code: "logout"})
            });

            if (!response.ok) {
                const data = await response.json();
                await Swal.fire("Error", data.output || "Error desconocido", "error");
                return;
            }

            const data = await response.json();

            if (data.output && data.output.split(":")[0] === "Error") {
                await Swal.fire("Error", data.output, "error");
                return;
            }

            await Swal.fire("Sesión cerrada", data.output || "Se ha cerrado la sesión con éxito.", "success");
        } catch (error) {
            Swal.fire("Error", "No se pudo conectar al servidor", "error");
        }
    };

    return (
        <nav className="navbar navbar-expand-lg navbar-light bg-light">
            <div className="container-fluid">
                <Link className="navbar-brand" to="/">
                    Mi Aplicación
                </Link>
                <div className="collapse navbar-collapse">
                    <ul className="navbar-nav me-auto mb-2 mb-lg-0">
                        <li className="nav-item">
                            <Link className="nav-link" to="/execution">
                                Ejecución
                            </Link>
                        </li>
                        <li className="nav-item">
                            <Link className="nav-link" to="/login">
                                Iniciar sesión
                            </Link>
                        </li>
                        <li className="nav-item">
                            <Link className="nav-link" to="/visualizador">Visualizador</Link>
                        </li>
                    </ul>
                    <div className="d-flex">
                        <button className="btn btn-outline-danger" onClick={handleLogout}>
                            Cerrar sesión
                        </button>
                    </div>
                </div>
            </div>
        </nav>
    );
};

export default Navbar;