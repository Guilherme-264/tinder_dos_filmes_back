CREATE DATABASE tinder_filmes;

\c tinder_filmes

CREATE TABLE salas (
  id VARCHAR(8) PRIMARY KEY,
  generos INTEGER[],
  streamings INTEGER[],
  status VARCHAR(20) DEFAULT 'lobby',
  criado_em TIMESTAMP DEFAULT NOW()
);