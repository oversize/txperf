/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/cardano-foundation/txperf/app"
	"github.com/spf13/cobra"
)

// walletCmd represents the wallet command
var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: runWallet,
}

func init() {

	rootCmd.AddCommand(walletCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// walletCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	walletCmd.Flags().StringVar(&walletName, "name", "", "Wallet name")
	walletCmd.MarkFlagRequired("name")
}

func runWallet(cmd *cobra.Command, args []string) {
	log.Default().Println("Run Wallet ")
	err := app.GetApp().CreateNewWallet(walletName)
	if err != nil {
		log.Fatalf("unable to create wallet: %s", err.Error())
	}
}
