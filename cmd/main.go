package main

import "os"

func main() {
	app := NewBlockchainApp()

	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}
