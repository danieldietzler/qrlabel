package cmd

import (
	"fmt"
	"github.com/danieldietzler/qrlabel/pdf"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "qrlabel [flags] <output file>",
	Short: "A tool for generating QR code labels",
	Long:  "A tool for generating QR code labels",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := pdf.CreatePdf(
			pdf.PageLayout{
				Cell: pdf.Cell{Width: 38, Height: 21.2}, Rows: 13, Cols: 5, Unit: "mm", SizeStr: "A4",
				LabelOrientation: pdf.LabelOrientation.BOTTOM,
			}, args[0], []pdf.Label{{"foo", "bar"}},
		)
		fmt.Println(file.Name())
	},
}

func init() {
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error executing command")
	}
}
