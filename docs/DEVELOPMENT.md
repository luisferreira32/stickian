# Local Development

- [Tools](#tools)
- [Run it locally: Docker Compose](#run-it-locally-docker-compose)
- [Run it locally: "Bare metal"](#run-it-locally-bare-metal)
- [Database migrations](#database-migrations)

## Tools

The techstack of the project is a Go backend, a PostgreSQL database and a React frontend. It is advised to use docker for reproducibility. So you'll need:

- Go, check required version in [go.mod](./go.mod)
- Node, check required version in [package.json](./package.json)
- Pnpm, for package management with the latest compatible version
- Prettier, for formatting the front-end code
- Docker, to spin-up the local development setup

## Run it locally: Docker Compose

This is the **preferred** way of development if you want a one-stop shop setup just working.

Start the local stack with Docker Compose:

```bash
docker compose up -w
```

Database frontend (PGAdmin):

- http://localhost:5050
- master password: 'admin'

Database (PostgreSQL):

- http://localhost:5432
- username/password: 'postgres'

Backend (Go server):

- http://localhost:8080

Frontend (React application):

- http://localhost:5173

Any changes to the source files React application and Go server will be reflected on the local development setup watch statements in [docker-compose.yml](../docker-compose.yml).

For a clean slate after development, remove everything with:

```bash
docker compose down -v
```

Some caveats of the approach:

- Any change to dependencies will rebuild the docker images, that will not re-use already downloaded dependencies and will take longer the more dependencies there are;
- Since Go is a compiled language, any changes to the server will rebuild the server image and require a re-compilation. This takes longer since cached compilation artifacts are not available on the docker image.

## Run it locally: "Bare metal"

A second option is to run it locally with a hybrid setup. This is to avoid the caveats above mentioned, but adds complexity to the setup.

The **one time** setup, and whenever dependencies change, is to download such dependencies:

```bash
pnpm install
go mod download
```

Then, start the database on a first terminal:

```bash
docker compose up postgres
```

Run the server on another terminal, and re-start the command whenever you need to test a new version of the code:

```bash
cd server && go run .
```

On a third terminal, run the web application, which will automatically do a hot-reload for any file changes:

```bash
pnpm dev
```

If you want to inspect the database, you can use `psql` with the dummy local database:

```bash
psql postgres://postgres:postgres@localhost:5432/app?sslmode=disable
```

# Database migrations

The database schema is defined under [migrations](../server/migrations/) and ran with [golang-migrate](https://github.com/golang-migrate/migrate) during the server startup. The tool performs migrations in a conservative way where SQL errors might set the migration state to dirty and avoid future actions until a human intervention, read more about it on the tool. This means, for development, you might reach an inconsistent state that needs to be fixed. This section of the documentation is to help you in such situations!

If your migrations are in a "dirty" state you should:

1. Fix your SQL! Otherwise, re-applying it will make the migration turn into a _dirty_ state again
2. Remove the dirty flag and reset the migration to the latest one run

Doing point 2. requires fiddling with the database data. Since in development you can run everything from a clean slate you can clean everything and re-start your development:

```bash
docker compose down -v
docker compose up -w
```

If the purpose is to not lose local test data, you have to connect to the database and rever the `schema_migrations` table up to the last applied version with `dirty` set to `0`.
