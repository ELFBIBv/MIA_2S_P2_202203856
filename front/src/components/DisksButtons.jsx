import React from "react";

const DisksButtons = ({disks, fetchPartitions,getDiskName }) => {
    return (

        <div className="row">
            {(() => {
                if (disks.length > 0) {
                    return disks.map((disk, index) => (
                        <div key={index} className="col-auto"> {/* col-auto ajusta el tamaño automático */}
                            <div
                                className="custom-card" /* Clase personalizada */
                                style={{cursor: "pointer", padding: "10px"}} /* Reducir padding */
                                onClick={() => fetchPartitions(disk)}
                            >
                                <div id="DiskImage" className="p-1"> {/* Ajustar padding de la imagen */}
                                </div>
                                <div>
                                    <h5 className="card-title text-center" style={{whiteSpace: "nowrap"}}>
                                        {getDiskName(disk)}
                                    </h5>
                                </div>
                            </div>
                        </div>
                    ));
                } else {
                    return <p>No hay discos con particiones montadas</p>;
                }
            })()}
        </div>
    );
};

export default DisksButtons;