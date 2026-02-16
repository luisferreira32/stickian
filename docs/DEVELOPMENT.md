# Local Development

- [Tools](#tools)
- [Run it locally: Docker Compose](#run-it-locally-docker-compose)
- [Run it locally: "Bare metal"](#run-it-locally-bare-metal)

## Tools

The techstack of the project is a Go backend, a PostgreSQL database and a React frontend. It is advised to use docker for reproducibility. So you'll need:

- Go, check required version in [go.mod](./go.mod)
- Node, check required version in [package.json](./package.json)
- Pnpm, for package management with the latest compatible version
- Prettier, for formatting the front-end code
- Docker, to spin-up the local development setup

## Run it locally: Docker Compose

Start the local stack with Docker Compose:

```bash
docker compose up
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

Any changes to the source files React application will be reflected on the local development setup due to the local mount in [docker-compose.yml](../docker-compose.yml).

To test server changes you need to re-build the server:

```bash
docker compose up --build server -d
```

## Run it locally: "Bare metal"

A second option is to run it without docker.

To run the server:

```bash
go run ./server/
```

To run the web application:

```bash
pnpm dev
```

_TODO: PostgreSQL setup_
