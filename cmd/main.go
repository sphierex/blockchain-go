package main

import (
	"log"

	"github.com/sphierex/blockchain-go/cmd/app"
)

func main() {
	a := app.New()

	if err := a.Execute(); err != nil {
		log.Println(err)
	}
}
