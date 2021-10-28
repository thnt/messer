# Simple App for Messer

## Development

### Server (Golang)

```bash
$ go mod tidy
$ go run main.go
```

### Client (Svelte)

```bash
$ yarn
$ yarn dev
```

## Deployment

### Build

```bash
$ yarn
$ yarn build
$ CGO_ENABLED=0 go build
```

### Config

View .env