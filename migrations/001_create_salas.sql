CREATE TABLE IF NOT EXISTS salas (
    id          VARCHAR(10) PRIMARY KEY,
    generos     INTEGER[],
    streamings  INTEGER[],
    criado_em   TIMESTAMP,
    status      VARCHAR(20)
);