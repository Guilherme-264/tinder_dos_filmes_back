CREATE TABLE IF NOT EXISTS salas (
    id          VARCHAR(10) PRIMARY KEY,
    generos     INTEGER[],
    streamings  INTEGER[],
    ano_inicio  INTEGER,
    ano_fim     INTEGER,
    nota_minima NUMERIC(3,1),
    diretor     VARCHAR(255),
    criado_em   TIMESTAMP,
    status      VARCHAR(20)
);
