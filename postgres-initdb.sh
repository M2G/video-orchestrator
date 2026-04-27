#!/bin/sh -e

psql --variable=ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
  CREATE DATABASE "postgres";
EOSQL

psql --variable=ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
  CREATE DATABASE "postgres_test";
EOSQL
