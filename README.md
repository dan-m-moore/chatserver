# chatserver

Welcome to **chatserver**, a multi-client chat server written in Go that serves both telnet and a web client.

## Build

Build `go build -o build/chatserver`

Test `go test -a -count=1 ./...`

Lint `go run github.com/golangci/golangci-lint/cmd/golangci-lint run --enable=unconvert --enable=dupl --enable=goconst --enable=gocyclo --enable=goimports --enable=maligned --enable=gochecknoinits --enable=gochecknoglobals --tests=false ./...`

Coverage `go test -a -count=1 -coverprofile build/coverage.out ./...`

Coverage HTML `go tool cover -html build/coverage.out -o build/coverage.html`

NOTE: Tested on Linux/OSX.  Requires Go Modules.

## Configuration/Usage

Config (located in `config.txt`)

- TelnetPort - the port to serve telnet on
- WebPort - the port to serve web client on
- WebClientPath - the location of the `webclient` dir
- LogFilePath - the location of the log file

Run `./build/chatserver -c config.txt`

Telnet Client `telnet localhost <TelnetPort>`

Web Client `http://localhost:<WebPort>`

## Backlog/Misc

Backlog:

- set up CI
- message deleting/editing
- authentication
- permissions
- direct messages/private channels
- modern web client
- switch from JSON RPC to gRPC or GraphQL
- model snapshots
- switch away from single threaded model (if performance requirements demand)

Cleanup:

- packages/functions that take filenames (config, model/actions) should use an os abstraction for easier unit testing
- action replayer/logger should be refactored for cleaner file I/O and more robust error handling
- subs engine should be extended to allow subscription rates and update type coalescing

Third Party Packages:

- github.com/reiver/go-oi
- github.com/reiver/go-telnet