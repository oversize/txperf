/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"log"

	"github.com/cardano-foundation/txperf/app"
	"github.com/cardano-foundation/txperf/cmd"
)

func main() {
	err := app.LoadApp()
	if err != nil {
		log.Fatalf("could not load app: %s", err)
	}
	cmd.Execute()
}
