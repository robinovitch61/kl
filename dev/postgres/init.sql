CREATE DATABASE flaskdb;
CREATE USER flaskuser WITH PASSWORD 'flaskpassword';
GRANT ALL PRIVILEGES ON DATABASE flaskdb TO flaskuser;

\connect flaskdb;

CREATE TABLE IF NOT EXISTS status_hits (
                                           id SERIAL PRIMARY KEY,
                                           hits INTEGER NOT NULL
);

INSERT INTO status_hits (hits)
SELECT 0
    WHERE NOT EXISTS (SELECT 1 FROM status_hits WHERE id = 1);
