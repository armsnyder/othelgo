# Othelgo

Compile with `make` (requires [Go](https://golang.org/doc/install) and [Golangci-lint](https://golangci-lint.run/usage/install/#local-installation)).

```sh
$ make
```

Run precompiled client after running `make`.

```sh
$ ./bin/client
```

Or skip `make` and run with `go`.

```sh
$ go run ./cmd/client
```

The client logs to file named `othelgo.log`.
