# Introduction
Go-based backend API for simple money transfer between accounts.
## Dependencies

To correctly run this server, ensure the following dependencies are satisfied:

- [Go](https://go.dev/doc/install)

## How to set up the project
1. Clone the repository
    ```bash
    serious@dev:~$ git clone git@github.com:ageeknamedslickback/simpleMoneyTransfer.git
    ```

2. Create `env.sh` and add the following environment variables
    ```bash
    export DB_USER=""
    export DB_PASS=""
    export DB_HOST=""
    export DB_PORT=""
    export DB_NAME=""
    export PORT=""
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
## How to run the APIs
1. API documentation

## How to run the tests

The server is covered by unit, integration and acceptance tests
```bash
serious@dev:~$ go test -v ./...
```

## API Spec

Export this collection to postman (if you are using it) to run the APIs:

[![Run in Postman](https://run.pstmn.io/button.svg)](https://app.getpostman.com/run-collection/a9495f127e246b807b17?action=collection%2Fimport)