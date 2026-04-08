CREATE TYPE biome_type AS ENUM (
    'mountain',
    'plains',
    'coast'
);


CREATE TABLE IF NOT EXISTS city (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id   UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(128)  NOT NULL,
    q           INT           NOT NULL,
    r           INT           NOT NULL,
    biome       biome_type    NOT NULL, 
    points      INT           NOT NULL DEFAULT 0,

    CONSTRAINT city_unique_coords UNIQUE (q, r)
);


CREATE TABLE IF NOT EXISTS city_resources (
    city_id         UUID        NOT NULL PRIMARY KEY REFERENCES city(id) ON DELETE CASCADE,
    food            INT         NOT NULL DEFAULT 0 CHECK (food >= 0),
    sticks          INT         NOT NULL DEFAULT 0 CHECK (sticks >= 0),
    stones          INT         NOT NULL DEFAULT 0 CHECK (stones >= 0),
    gems            INT         NOT NULL DEFAULT 0 CHECK (gems >= 0),
    population      INT         NOT NULL DEFAULT 0 CHECK (population >= 0),
    faith           INT         NOT NULL DEFAULT 0 CHECK (faith >= 0)
);


CREATE TABLE IF NOT EXISTS city_buildings (
    city_id         UUID        NOT NULL PRIMARY KEY REFERENCES city(id) ON DELETE CASCADE,
    city_hall       INT         NOT NULL DEFAULT 0,
    embassy         INT         NOT NULL DEFAULT 0,
    treasury        INT         NOT NULL DEFAULT 0, 
    tavern          INT         NOT NULL DEFAULT 0,
    farm            INT         NOT NULL DEFAULT 0,
    lumbermill      INT         NOT NULL DEFAULT 0,
    quarry          INT         NOT NULL DEFAULT 0,
    crystal_mine    INT         NOT NULL DEFAULT 0,
    warehouse       INT         NOT NULL DEFAULT 0,
    market          INT         NOT NULL DEFAULT 0,
    harbor          INT         NOT NULL DEFAULT 0,
    walls           INT         NOT NULL DEFAULT 0,
    barracks        INT         NOT NULL DEFAULT 0,
    docks           INT         NOT NULL DEFAULT 0,
    spy_guild       INT         NOT NULL DEFAULT 0,
    library         INT         NOT NULL DEFAULT 0,
    workshop        INT         NOT NULL DEFAULT 0,
    observatory     INT         NOT NULL DEFAULT 0,
    temple          INT         NOT NULL DEFAULT 0,
    shrine          INT         NOT NULL DEFAULT 0,
    cathedral       INT         NOT NULL DEFAULT 0
);
