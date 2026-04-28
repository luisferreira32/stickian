CREATE TABLE IF NOT EXISTS world (
    q INTEGER NOT NULL,
    r INTEGER NOT NULL,
    settleable BOOLEAN NOT NULL,
    biome INTEGER NOT NULL,
    PRIMARY KEY (q, r)
);
