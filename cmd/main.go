package main

import (
	"os"

	"github.com/sphierex/blockchain-go/cmd/handlers"
)

func main() {
	app := handlers.NewBlockchainApp()

	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}
