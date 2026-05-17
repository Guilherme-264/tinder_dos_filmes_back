CREATE TABLE IF NOT EXISTS filmes (
    id              SERIAL PRIMARY KEY,
    tmdb_id         INTEGER UNIQUE NOT NULL,
    titulo          VARCHAR(255),
    visao_geral     TEXT,
    poster_path     VARCHAR(255),
    nota_media      NUMERIC(3,1),
    data_lancamento VARCHAR(20),
    generos         INTEGER[],
    streamings      INTEGER[]
);