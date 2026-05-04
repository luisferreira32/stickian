import os
import json

import psycopg2
from decouple import config


def load_data():
    with open(os.path.join("world_data", "world.csv"), "r") as f:
        data = f.readlines()

    with open(os.path.join("world_data", "world_settleable.json"), "r") as f:
        settleable = json.load(f)

    return data, settleable


def write_to_db(data, settleable):
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
            AND table_name = 'world'
        );
    """)
    if not cursor.fetchone()[0]:
        cursor.close()
        conn.close()
        raise ValueError("⛔ Table 'world' does not exist. Please create it first.")

    # verify if table is empty
    cursor.execute("SELECT EXISTS (SELECT 1 FROM world);")
    if cursor.fetchone()[0]:
        cursor.close()
        conn.close()
        raise ValueError("⛔ Table 'world' is not empty. Please truncate it first.")

    for q, line in enumerate(data):
        for r, biome in enumerate(line.split(",")[:-1]):
            is_settleable = "true" if [q, r] in settleable else "false"
            cursor.execute(
                "INSERT INTO world (q, r, settleable, biome) VALUES (%s, %s, %s, %s)", (q, r, is_settleable, biome)
            )
    conn.commit()
    print("✅ World map data successfully inserted into database")

    cursor.close()
    conn.close()


def run():
    data, settleable = load_data()
    write_to_db(data, settleable)


if __name__ == "__main__":
    run()
