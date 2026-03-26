import os
from decouple import config
import psycopg2


def load_data():
    with open(os.path.join("world_data", "world.csv"), "r") as f:
        data = f.readlines()
    return data


def write_to_db(data):
    try:
        conn = psycopg2.connect(database = config("DATABASE_NAME"),
                                user = config("DATABASE_USER"),
                                password = config("DATABASE_PASSWORD"),
                                host = config("DATABASE_HOST"),
                                port = config("DATABASE_PORT"))
    except Exception as e:
        print(f"Error: {e}")
        return
    cursor = conn.cursor()
    for q, line in enumerate(data):
        for r, biome in enumerate(line.split(",")[:-1]):
            values = (q, r, biome)
            cursor.execute("INSERT INTO world (q, r, biome) VALUES (%s, %s, %s)", values)
    conn.commit()
    cursor.close()
    conn.close()


def run():
    data = load_data()
    write_to_db(data)


if __name__ == "__main__":
    run()
