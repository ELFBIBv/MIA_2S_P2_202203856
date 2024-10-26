const CommandService = {
    // Función para procesar un comando y enviar al backend
    parseAndSend: async (command) => {
        const parsedCommand = CommandService.parseCommand(command);
        if (parsedCommand) {
            const result = await CommandService.sendCommand(parsedCommand);

            // Detectar si el comando es mkdisk y almacenar el path del disco en localStorage
            if (command.startsWith("mkdisk")) {
                const disks = JSON.parse(localStorage.getItem("disks")) || [];
                disks.push(parsedCommand.body.path);  // Almacenar solo el path del disco creado
                localStorage.setItem("disks", JSON.stringify(disks));
            }

            return result;
        } else {
            throw new Error("Comando no válido o no soportado");
        }
    },

    // Función para enviar los comandos al backend con mejor manejo de errores
    sendCommand: async (parsedCommand) => {
        try {
            const response = await fetch(parsedCommand.url, {
                method: parsedCommand.method,
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify(parsedCommand.body)
            });

            // Si el estado no es ok, arroja un error con el contenido de la respuesta
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || `Error en el servidor: ${response.statusText}`);
            }

            // Devolver la respuesta en JSON si todo está bien
            return await response.json();
        } catch (error) {
            // Capturar cualquier error durante la solicitud
            throw new Error(error.message || "Error al enviar el comando");
        }
    }
};

export default CommandService;