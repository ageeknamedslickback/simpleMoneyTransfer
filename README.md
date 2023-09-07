# Money Transfer
## Introduction
Go-based backend API for simple money transfer between accounts.
## Dependencies

To correctly run this server, ensure the following dependencies are satisfied:

- [Go](https://go.dev/doc/install)
- PostgreSQL
- Auth0

## How it all works

To understand how everything comes together, here's a Confluence page that gives more context:

- [Confluence](https://kenmathengendungu.atlassian.net/l/cp/ez1K4gda)

## How to set up the project
1. Clone the repository
    ```bash
    serious@dev:~$ git clone git@github.com:ageeknamedslickback/simpleMoneyTransfer.git
    ```

2. Create `env.sh` and add the following environment variables. This assumes that you have
created a database whose information is populated under `DB_` prefix.
    ```bash
    export DB_USER=""
    export DB_PASS=""
    export DB_HOST=""
    export DB_PORT=""
    export DB_NAME=""
    export PORT=""
    export PORT=8080
    export AUTH0_GRANT_TYPE=""
    export AUTH0_CLIENT_ID=""
    export AUTH0_CLIENT_SECRET=""
    export AUTH0_AUDIENCE=""
    export AUTH0_DOMAIN=""
    ```

3. Install Go dependencies
    ```bash
    serious@dev:~$ go mod tidy
    ```

4. Run the server (performs the migrations to your database)
    ```bash
    serious@dev:~$ source env.sh
    serious@dev:~$ go run server.go
    ```
## How to run the tests

The server is covered by unit, integration and acceptance tests
```bash
serious@dev:~$ go test -v ./...
```

## API Spec

Export this collection to postman (if you are using it) to run the APIs:

[![Run in Postman](https://run.pstmn.io/button.svg)](https://api.postman.com/collections/7960412-d8519d9e-e031-41fa-9c3a-74660876ab2b?access_key=PMAT-01H9Q4R2AQAF7YJ86G1P8BV4CP)

## Developer

Kenneth Mathenge | ken.mathenge.ndungu@gmail.com