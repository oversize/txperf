/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/cardano-foundation/txperf/app"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run txperf",
	Long:  `Run txperf`,
	RunE:  run,
}

var name string

func init() {
	rootCmd.AddCommand(runCmd)
}

func run(cmd *cobra.Command, args []string) error {
	log.Default().Println("Run Run Command")
	err := app.GetApp().Run()
	if err != nil {
		return err
	}
	return nil
}
