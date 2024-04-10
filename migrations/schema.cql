CREATE KEYSPACE scyllatest WITH REPLICATION = { 'class': 'NetworkTopologyStrategy', 'replication_factor' : 3 }
                            AND DURABLE_WRITES = TRUE;

USE scyllatest;

CREATE TABLE randomdata (
    id timeuuid PRIMARY KEY,
    data int,
);
