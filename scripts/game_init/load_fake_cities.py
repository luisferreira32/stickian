import psycopg2
from decouple import config
import uuid

admin_id = "fed6d389-b09c-4ff1-afa6-215ca3dcd449"

fake_cities = {
    str(uuid.uuid4()): {
        "player_id": admin_id,
        "name": "Big Stickland",
        "q": 50,
        "r": 50,
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
        },
    },
    str(uuid.uuid4()): {
        "player_id": admin_id,
        "name": "Wowsticks",
        "q": 51,
        "r": 49,
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
        "q": 53,
        "r": 52,
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
        "q": 47,
        "r": 48,
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
    },
    str(uuid.uuid4()): {
        "player_id": admin_id,
        "name": "Ironhold",
        "q": 49,
        "r": 54,
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
        return (
            city_id,
            city["resources"]["food"] if "food" in city["resources"] else 0,
            city["resources"]["sticks"] if "sticks" in city["resources"] else 0,
            city["resources"]["stones"] if "stones" in city["resources"] else 0,
            city["resources"]["gems"] if "gems" in city["resources"] else 0,
            city["resources"]["population"] if "population" in city["resources"] else 0,
            city["resources"]["faith"] if "faith" in city["resources"] else 0,
        )
    elif keyword == "city_buildings":
        return (
            city_id,
            city["buildings"]["city_hall"] if "city_hall" in city["buildings"] else 0,
            city["buildings"]["embassy"] if "embassy" in city["buildings"] else 0,
            city["buildings"]["treasury"] if "treasury" in city["buildings"] else 0,
            city["buildings"]["tavern"] if "tavern" in city["buildings"] else 0,
            city["buildings"]["farm"] if "farm" in city["buildings"] else 0,
            city["buildings"]["lumbermill"] if "lumbermill" in city["buildings"] else 0,
            city["buildings"]["quarry"] if "quarry" in city["buildings"] else 0,
            city["buildings"]["crystal_mine"] if "crystal_mine" in city["buildings"] else 0,
            city["buildings"]["warehouse"] if "warehouse" in city["buildings"] else 0,
            city["buildings"]["market"] if "market" in city["buildings"] else 0,
            city["buildings"]["harbor"] if "harbor" in city["buildings"] else 0,
            city["buildings"]["walls"] if "walls" in city["buildings"] else 0,
            city["buildings"]["barracks"] if "barracks" in city["buildings"] else 0,
            city["buildings"]["docks"] if "docks" in city["buildings"] else 0,
            city["buildings"]["spy_guild"] if "spy_guild" in city["buildings"] else 0,
            city["buildings"]["library"] if "library" in city["buildings"] else 0,
            city["buildings"]["workshop"] if "workshop" in city["buildings"] else 0,
            city["buildings"]["observatory"] if "observatory" in city["buildings"] else 0,
            city["buildings"]["temple"] if "temple" in city["buildings"] else 0,
            city["buildings"]["shrine"] if "shrine" in city["buildings"] else 0,
            city["buildings"]["cathedral"] if "cathedral" in city["buildings"] else 0,
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

    for id, city in fake_cities.items():
        insert_city(cursor, city, id)
        insert_city_resources(cursor, city, id)
        insert_city_buildings(cursor, city, id)
    conn.commit()
    print("✅ Cities data successfully inserted into databases")

    cursor.close()
    conn.close()


def run():
    write_to_db(fake_cities)


if __name__ == "__main__":
    run()
