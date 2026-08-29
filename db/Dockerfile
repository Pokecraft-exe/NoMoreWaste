# Base de donnees NO MORE WASTE (PostgreSQL). A construire depuis ce dossier,
# sql.sql et seed.sql y etant utilises comme scripts d'initialisation :
#   docker build -t nomorewaste-db .
FROM postgres:16-alpine

ENV POSTGRES_DB=nomorewaste \
    POSTGRES_USER=postgres \
    POSTGRES_PASSWORD=nomorewaste

# Executes automatiquement au tout premier demarrage (volume de donnees
# vide), dans l'ordre alphabetique, comme documente par l'image officielle
# postgres.
COPY sql.sql /docker-entrypoint-initdb.d/01-schema.sql
COPY seed.sql /docker-entrypoint-initdb.d/02-seed.sql

EXPOSE 5432
