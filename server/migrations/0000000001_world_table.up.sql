CREATE TABLE IF NOT EXISTS world (
    q           INT           NOT NULL,
    r           INT           NOT NULL,
    biome       INT           NOT NULL,
    settleable  BOOLEAN       NOT NULL,
    PRIMARY KEY (q, r)
);
