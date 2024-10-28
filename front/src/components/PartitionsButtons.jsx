import React from "react";

const PartitionsButtons = ({partitions, fetchFilesSystem }) => {
    return (
        <div className="row">
            {(() => {
                if (partitions.length > 0) {
                    return partitions.map((partition, index) => (
                        <div key={index} className="col-auto"> {/* col-auto ajusta el tamaño automático */}
                            <div
                                className="custom-card" /* Clase personalizada */
                                style={{cursor: "pointer", padding: "10px"}} /* Reducir padding */
                                onClick={() => fetchFilesSystem(partition.Name)}
                            >
                                <div id="PartitionImage" className="p-1"> {/* Ajustar padding de la imagen */}
                                </div>
                                <div>
                                    <h5 className="card-title text-center" style={{whiteSpace: "nowrap"}}>
                                        {partition.Name}
                                    </h5>
                                </div>
                            </div>
                        </div>
                    ));
                } else {
                    return <p>No se han creado discos aún.</p>;
                }
            })()}
        </div>
    );
};

export default PartitionsButtons;