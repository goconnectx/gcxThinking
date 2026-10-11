# AGENTS.md

## Project Overview

A minimal Go program that sends a single HTTP POST request to a local "SystemOne" endpoint (served at `http://localhost:11434`) and prints the raw response body. It asks the `tev1` model to evaluate a piece of text (`state`) against a set of yes/no style questions.

There is one file, `main.go`, and no other dependencies beyond the Go standard library.

## Setup

- Language: Go (standard library only: `fmt`, `io/ioutil`, `log`, `net/http`, `strings`)
- A server must be listening at `http://localhost:11434` and expose `POST /v1/systemone`.
- The model `tev1` must be available on that server.

## Commands

```bash
go run main.go     # run the request and print the response
go build -o app    # build a binary
go vet ./...       # static checks
gofmt -l .         # check formatting
