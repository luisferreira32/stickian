import psycopg2
from decouple import config
import uuid
import random

admin_id = "3c6ee013-6db7-458e-865e-5c1ba8375688"

fake_cities = {
    str(uuid.uuid4()): {
        "player_id": admin_id,
        "name": "Big Stickland",
        "q": 68,
        "r": 14,
        "biome": "plains",
        "points": 350,
        "buildings": {
            "city_hall": 4,
            "farm": 2,
            "quarry": 2,
            "lumbermill": 2,
            "crystal_mine": 3,
            "market": 1,
            "warehouse": 1,
        },
        "resources": {
            "food": 45,
            "sticks": 312,
            "stones": 215,
            "gems": 145,
            "population": 45,
            "faith": 18,
        }
    },
    str(uuid.uuid4()): {
        "player_id": admin_id,
        "name": "Wowsticks",
        "q": 24,
        "r": 73,
        "biome": "mountain",
        "points": 210,
        "buildings": {
            "city_hall": 3,
            "quarry": 4,
            "crystal_mine": 5,
            "walls": 2,
            "barracks": 1,
        },
        "resources": {
            "food": 28,
            "sticks": 80,
            "stones": 640,
            "gems": 310,
            "population": 28,
            "faith": 5,
        },
    },
    str(uuid.uuid4()): {
        "player_id": admin_id,
        "name": "Pila Cave",
        "q": 8,
        "r": 14,
        "biome": "coast",
        "points": 480,
        "buildings": {
            "city_hall": 6,
            "farm": 5,
            "quarry": 5,
            "lumbermill": 5,
            "crystal_mine": 5,
            "harbor": 1,
            "docks": 2,
            "market": 2,
            "tavern": 1,
            "embassy": 1,
        },
        "resources": {
            "food": 72,
            "sticks": 150,
            "stones": 90,
            "gems": 60,
            "population": 890,
            "faith": 42,
        },
    },
    str(uuid.uuid4()): {
        "player_id": admin_id,
        "name": "Verdant Vale",
        "q": 25,
        "r": 26,
        "biome": "plains",
        "points": 125,
        "buildings": {
            "city_hall": 2,
            "farm": 3,
            "lumbermill": 3,
            "shrine": 1,
        },
        "resources": {
            "food": 18,
            "sticks": 420,
            "stones": 55,
            "gems": 10,
            "population": 750,
            "faith": 88,
        },
        "units": {
            "swordsmen": 200,
            "archers": 200,
            "cavalry": 100,
            "ships": 20,
            "spies": 5,
        },
    },
    str(uuid.uuid4()): {
        "player_id": admin_id,
        "name": "Ironhold",
        "q": 96,
        "r": 42,
        "biome": "mountain",
        "points": 560,
        "buildings": {
            "city_hall": 6,
            "quarry": 5,
            "walls": 4,
            "barracks": 3,
            "workshop": 2,  
            "observatory": 1,
        },
        "resources": {
            "food": 60,
            "sticks": 200,
            "stones": 980,
            "gems": 500,
            "population": 300,
            "faith": 12,
        },
        "units": {
            "swordsmen": 100,
            "archers": 100,
            "cavalry": 10,
            "ships": 0,
            "spies": 0,
        },
    },
}


def serialize_data(city, keyword, city_id):
    if keyword == "city":
        return (
            city_id,
            city["player_id"],
            city["name"],
            city["q"],
            city["r"],
            city["biome"],
            city["points"],
        )
    elif keyword == "city_resources":
        res = city.get("resources", {})
        return (
            city_id,
            res.get("food", 0),
            res.get("sticks", 0),
            res.get("stones", 0),
            res.get("gems", 0),
            res.get("population", 0),
            res.get("faith", 0),
        )
    elif keyword == "city_buildings":
        bld = city.get("buildings", {})
        return (
            city_id,
            bld.get("city_hall", 0),
            bld.get("embassy", 0),
            bld.get("treasury", 0),
            bld.get("tavern", 0),
            bld.get("farm", 0),
            bld.get("lumbermill", 0),
            bld.get("quarry", 0),
            bld.get("crystal_mine", 0),
            bld.get("warehouse", 0),
            bld.get("market", 0),
            bld.get("harbor", 0),
            bld.get("walls", 0),
            bld.get("barracks", 0),
            bld.get("docks", 0),
            bld.get("spy_guild", 0),
            bld.get("library", 0),
            bld.get("workshop", 0),
            bld.get("observatory", 0),
            bld.get("temple", 0),
            bld.get("shrine", 0),
            bld.get("cathedral", 0),
        )
    elif keyword == "city_troops":
        units = city.get("units", {})
        return (
            city_id,
            units.get("swordsmen", 0),
            units.get("archers", 0),
            units.get("cavalry", 0),
            units.get("ships", 0),
            units.get("spies", 0),
        )


def insert_city(cursor, city, city_id):
    cursor.execute(
        "INSERT INTO city (id, player_id, name, q, r, biome, points) VALUES (%s, %s, %s, %s, %s, %s, %s)",
        serialize_data(city, "city", city_id),
    )


def insert_city_resources(cursor, city, city_id):
    cursor.execute(
        "INSERT INTO city_resources (city_id, food, sticks, stones, gems, population, faith) VALUES (%s, %s, %s, %s, %s, %s, %s)",
        serialize_data(city, "city_resources", city_id),
    )


def insert_city_buildings(cursor, city, city_id):
    cursor.execute(
        "INSERT INTO city_buildings (city_id, city_hall, embassy, treasury, tavern, farm, lumbermill, quarry, crystal_mine, warehouse, market, harbor, walls, barracks, docks, spy_guild, library, workshop, observatory, temple, shrine, cathedral) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)",
        serialize_data(city, "city_buildings", city_id),
    )


def insert_city_troops(cursor, city, city_id):
    cursor.execute(
        "INSERT INTO city_troops (city_id, swordsmen, archers, cavalry, ships, spies) VALUES (%s, %s, %s, %s, %s, %s)",
        serialize_data(city, "city_troops", city_id),
    )


def write_to_db(data):
    try:
        conn = psycopg2.connect(
            database=config("DATABASE_NAME"),
            user=config("DATABASE_USER"),
            password=config("DATABASE_PASSWORD"),
            host=config("DATABASE_HOST"),
            port=config("DATABASE_PORT"),
        )
    except Exception as e:
        raise ValueError(f"⛔ Error: {e}")
    cursor = conn.cursor()

    # verify if table exists
    cursor.execute("""
        SELECT EXISTS (
            SELECT 1
            FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = 'city'
        );
    """)
    if not cursor.fetchone()[0]:
        cursor.close()
        conn.close()
        raise ValueError("⛔ Table 'city' does not exist. Please create it first.")

    cursor.execute("SELECT q, r, biome FROM world WHERE settleable = 'true';")
    settleable_tiles = cursor.fetchall()



    if len(settleable_tiles) < len(fake_cities):
        raise ValueError("Not enough settleable tiles for fake cities.")

    sampled_tiles = random.sample(settleable_tiles, len(fake_cities))

    for (city_id, city), tile in zip(fake_cities.items(), sampled_tiles):
        city["q"] = tile[0]
        city["r"] = tile[1]
        city["biome"] = tile[2]
        insert_city(cursor, city, city_id)
        insert_city_resources(cursor, city, city_id)
        insert_city_buildings(cursor, city, city_id)
        insert_city_troops(cursor, city, city_id)
    conn.commit()
    print("✅ Cities data successfully inserted into databases")

    cursor.close()
    conn.close()


def run():
    write_to_db(fake_cities)


if __name__ == "__main__":
    run()
