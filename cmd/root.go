package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "qrlabel [flags] <output file>",
	Short: "A tool for generating QR code labels",
	Long:  "A tool for generating QR code labels",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello world!")
	},
}

func init() {
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Unknown command")
	}
}
