# Othelgo

A commandline Othello game and experiment using AWS API Gateway WebSocket APIs and AWS Lambda.

## Play the game

Run the client with `make run` (requires [Go](https://golang.org/doc/install)).

```sh
$ make run
```

## Local development

Requires [Go](https://golang.org/doc/install) and [Docker Compose](https://docs.docker.com/compose/install/).

In one terminal window, start the local server with `make serve`.

```sh
$ make serve
```

In a second and third terminal window, start the client in local mode with `make playlocal`.

```sh
$ make playlocal
```
