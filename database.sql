CREATE DATABASE tinder_filmes;

\c tinder_filmes

CREATE TABLE salas (
  id VARCHAR(8) PRIMARY KEY,
  generos INTEGER[],
  streamings INTEGER[],
  ano_inicio INTEGER,
  ano_fim INTEGER,
  nota_minima NUMERIC(3,1),
  diretor VARCHAR(255),
  status VARCHAR(20) DEFAULT 'lobby',
  criado_em TIMESTAMP DEFAULT NOW()
);
