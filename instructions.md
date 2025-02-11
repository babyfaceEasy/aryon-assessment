### Setting Up Development Environment

1. Clone the repository

```sh
 git git@github.com:babyfaceEasy/aryon-assessment.git
 cd aryon-assessment
```

2. Setup your env variables: a copy of the required variables can be found in the `.env.example` file. copy and configure using your custom setup

```
 cp .env.example .env
```

3. Install Dependencies

```sh
 go mod tidy
```

4. Setup your database: the database specified in .env `POSTGRES_DATABASE` must be created before proceeding

5.  Run for development mode 

```
docker-compose -f docker-compose.dev.yml up -d
```

To Run in Production mode

```
docker-compose -f docker-compose.yml up -d
```

6.  You can make use of any gRPC client to access the needed endpoints on `http://localhost:50051`.
```
GetConnector
GetConnectors
SaveConnector
DeleteConnector
```
NB: The proto file is located inside the `protobuf` folder.

### Notes on Key functionalities

- We use slog for logging
- We use `PGXPool`  for Database access this is for speed and efficiency.
- DB migrations are very low level: **hand written atomic SQL commands** , managed using [golang migrate](https://github.com/golang-migrate/migrate) and are run automatically on server startup

### Deployments

The project is packaged using docker containers. a production grade `Dockerfile` already exists and be used to replicate remote environments locally