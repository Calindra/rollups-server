# Rollups Server

## Running the server

```
go run main.go
```

If you've built using go build command, run 

```
./rollups-server
```

## Other usage

To generate routes based on yaml files, run 

```
go generate ./...
```

or 

```
make gen
```

To build the application, run 

```
go build -o rollups-server
```

or 

```
make build
```

To execute tests, run 

```
go test -p 1 ./...
```

or 

```
make test
```