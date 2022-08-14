#!/bin/bash

# Below this comment script will insert current user name while building: user=$USER
# Below this comment script will insert user password while building: password=$USER_PASSWORD

psql -U postgres <<- EOSQL
    CREATE USER $user WITH PASSWORD '$password';
    CREATE DATABASE $user;
    GRANT ALL PRIVILEGES ON DATABASE $user TO $user;
    CREATE DATABASE "metrics";
    GRANT ALL PRIVILEGES ON DATABASE "metrics" TO $user;
EOSQL
