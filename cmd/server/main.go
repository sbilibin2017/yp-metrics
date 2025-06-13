package main

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	config := parseFlags()
	err := run(context.Background(), config)
	if err != nil {
		panic(err)
	}
}
