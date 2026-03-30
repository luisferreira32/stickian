CREATE TYPE biome_type AS ENUM (
    'mountain',
    'plains',
    'coast'
);


CREATE TABLE city (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id   UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(128)  NOT NULL,
    q           INT           NOT NULL,
    r           INT           NOT NULL,
    biome       biome_type    NOT NULL, 
    points      INT           NOT NULL DEFAULT 0,
 
    CONSTRAINT city_unique_coords UNIQUE (q, r)
);


CREATE TABLE city_resources (
    city_id         UUID        NOT NULL PRIMARY KEY REFERENCES city(id) ON DELETE CASCADE,
    food            INT         NOT NULL DEFAULT 0 CHECK (food >= 0),
    sticks          INT         NOT NULL DEFAULT 0 CHECK (sticks >= 0),
    rocks           INT         NOT NULL DEFAULT 0 CHECK (rocks >= 0),
    gems            INT         NOT NULL DEFAULT 0 CHECK (gems >= 0),
    population      INT         NOT NULL DEFAULT 0 CHECK (population >= 0),
    faith           INT         NOT NULL DEFAULT 0 CHECK (faith >= 0)
);

CREATE TABLE building (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    city_id           UUID          NOT NULL REFERENCES city(id) ON DELETE CASCADE,
    building_type     VARCHAR(64)   NOT NULL REFERENCES building_type(name),
    level             INT           NOT NULL DEFAULT 1 CHECK (level >= 1),
    points            INT           NOT NULL DEFAULT 0,
 
    CONSTRAINT building_unique_per_city UNIQUE (city_id, building_type)
);

CREATE TABLE building_type (
    name        VARCHAR(64) PRIMARY KEY,
    max_level   INT         NOT NULL DEFAULT 3 CHECK (max_level >= 1),
    description TEXT
);

CREATE TABLE building_level_stats (
    building_type_name  VARCHAR(64)     REFERENCES building_type(name),
    level               INT             CHECK (level >= 1),
    points              INT             NOT NULL DEFAULT 0,
    cost_sticks         INT             NOT NULL DEFAULT 0,
    cost_rocks          INT             NOT NULL DEFAULT 0,
    cost_gems           INT             NOT NULL DEFAULT 0,
    upgrade_duration    INT             NOT NULL DEFAULT 0,
    PRIMARY KEY (building_type_name, level)
);

CREATE TABLE building_bonus_type (
    name        VARCHAR(64) PRIMARY KEY,  -- 'yield_food', 'construction_speed', etc.
    description TEXT
);

CREATE TABLE building_level_bonus (
    building_type_name  VARCHAR(64)     NOT NULL REFERENCES building_type(name),
    level               INT             NOT NULL CHECK (level >= 1),
    bonus_type          VARCHAR(64)     NOT NULL REFERENCES building_bonus_type(name),
    value               NUMERIC         NOT NULL,
    PRIMARY KEY (building_type_name, level, bonus_type)
);