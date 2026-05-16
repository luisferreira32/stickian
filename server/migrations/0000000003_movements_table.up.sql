CREATE TABLE IF NOT EXISTS movement (
    id              UUID            NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    city_from       UUID            NOT NULL REFERENCES city(id) ON DELETE CASCADE,
    city_to         UUID            NOT NULL REFERENCES city(id),
    type            INT             NOT NULL,
    arrival_time    TIMESTAMPTZ     NOT NULL,

    -- troops
    swordsmen       INT             NOT NULL DEFAULT 0 CHECK (swordsmen >= 0),
    archers         INT             NOT NULL DEFAULT 0 CHECK (archers >= 0),
    cavalry         INT             NOT NULL DEFAULT 0 CHECK (cavalry >= 0),
    ships           INT             NOT NULL DEFAULT 0 CHECK (ships >= 0),
    spies           INT             NOT NULL DEFAULT 0 CHECK (spies >= 0),

    -- resources
    food            INT             NOT NULL DEFAULT 0 CHECK (food >= 0),
    sticks          INT             NOT NULL DEFAULT 0 CHECK (sticks >= 0),
    stones          INT             NOT NULL DEFAULT 0 CHECK (stones >= 0),
    gems            INT             NOT NULL DEFAULT 0 CHECK (gems >= 0),

    CONSTRAINT movement_different_cities CHECK (city_from != city_to)
);

CREATE INDEX IF NOT EXISTS movement_city_from_idx ON movement (city_from);
CREATE INDEX IF NOT EXISTS movement_city_to_idx ON movement (city_to);
