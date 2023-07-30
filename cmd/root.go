package cmd

import (
	"fmt"
	"github.com/danieldietzler/qrlabel/pdf"
	"github.com/spf13/cobra"
)

var (
	Height        float64
	Width         float64
	Rows          int
	Cols          int
	Unit          string
	SizeStr       string
	LabelPosition pdf.Position
)

var rootCmd = &cobra.Command{
	Use:   "qrlabel [flags] <output file>",
	Short: "A tool for generating QR code labels",
	Long:  "A tool for generating QR code labels",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file, err := pdf.CreatePdf(
			pdf.PageLayout{
				Cell:          pdf.Cell{Width: Width, Height: Height},
				Rows:          Rows,
				Cols:          Cols,
				Unit:          Unit,
				SizeStr:       SizeStr,
				LabelPosition: LabelPosition,
			}, args[0], []pdf.Label{{"foo", "bar"}},
		)
		fmt.Println(file.Name())
		return err
	},
}

func init() {
	rootCmd.Flags().Float64VarP(&Width, "width", "W", 38, "Width of a label in the specified unit")
	rootCmd.Flags().Float64VarP(
		&Height, "height", "H", 21.2, "Height of a label in the specified unit",
	)
	rootCmd.Flags().IntVarP(&Rows, "rows", "r", 10, "Number of rows per page")
	rootCmd.Flags().IntVarP(&Cols, "cols", "c", 5, "Number of columns per page")
	rootCmd.Flags().StringVarP(
		&Unit, "unit", "u", "mm", `Unit of measurement. Available options are "pt", "mm", "cm" and "inch"`,
	)
	rootCmd.Flags().StringVarP(
		&SizeStr, "size", "s", "A4",
		`The size of the output pdf page. Available options are "A3", "A4", "A5", "Letter" and "Legal"`,
	)
	rootCmd.Flags().StringVarP(
		(*string)(&LabelPosition), "position", "p", string(pdf.LabelPosition.LEFT),
		`Label position relative to the QR code. Available options are "T" (top), "B" (bottom), "L" (left), "R" (right)`,
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error executing command")
	}
}
