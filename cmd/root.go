package cmd

import (
	"fmt"
	"github.com/danieldietzler/qrlabel/pdf"
	"github.com/go-pdf/fpdf"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

var (
	LabelHeight   float64
	LabelWidth    float64
	Rows          int
	Cols          int
	Unit          string
	SizeStr       string
	PageHeight    float64
	PageWidth     float64
	LabelPosition pdf.Position
	InputFileName string
	Separator     string
	RecoveryLevel qrcode.RecoveryLevel
)

var rootCmd = &cobra.Command{
	Use:   "qrlabel [flags] <output file>",
	Short: "A tool for generating QR code labels",
	Long:  "A tool for generating QR code labels",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var inputReader io.Reader
		if len(InputFileName) == 0 {
			inputReader = cmd.InOrStdin()
		} else {
			var err error
			inputReader, err = os.Open(InputFileName)

			if err != nil {
				return err
			}
		}

		inputStream, err := io.ReadAll(inputReader)

		if err != nil {
			return err
		}

		input := string(inputStream)

		var labels []pdf.Label
		for _, line := range strings.Split(input, "\n") {
			if line == "" {
				continue
			}
			elements := strings.SplitN(line, Separator, 2)

			if len(elements) == 2 {
				labels = append(labels, pdf.Label{Content: elements[0], Label: elements[1]})
			} else if len(elements) == 1 {
				labels = append(labels, pdf.Label{Content: elements[0], Label: elements[0]})
			}

		}

		if cmd.Flags().Changed("pageHeight") {
			SizeStr = ""
		}

		file, err := pdf.CreatePdf(
			pdf.PageLayout{
				Cell:    pdf.Cell{Width: LabelWidth, Height: LabelHeight},
				Rows:    Rows,
				Cols:    Cols,
				Unit:    Unit,
				SizeStr: SizeStr,
				Size: fpdf.SizeType{
					Wd: PageWidth,
					Ht: PageHeight,
				},
				LabelPosition: LabelPosition,
			}, RecoveryLevel, args[0], labels,
		)
		fmt.Println(file.Name())
		return err
	},
}

func init() {
	rootCmd.AddCommand(CompletionCmd)

	rootCmd.Flags().Float64VarP(&LabelWidth, "width", "w", 38, "Width of a label in the specified unit")
	rootCmd.Flags().Float64VarP(
		&LabelHeight, "height", "H", 21.2, "Height of a label in the specified unit",
	)
	rootCmd.Flags().IntVarP(&Rows, "rows", "r", 10, "Number of rows per page")
	rootCmd.Flags().IntVarP(&Cols, "cols", "c", 5, "Number of columns per page")
	rootCmd.Flags().StringVarP(
		&Unit, "unit", "u", "mm", `Unit of measurement ["pt"|"mm"|"cm"|"inch"]`,
	)
	rootCmd.RegisterFlagCompletionFunc(
		"unit", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"pt", "mm", "cm", "inch"}, cobra.ShellCompDirectiveDefault
		},
	)
	rootCmd.Flags().StringVarP(
		&SizeStr, "pageSize", "s", "A4",
		`The size string of the output pdf page ["A3"|"A4"|"A5"|"Letter"|"Legal"]`,
	)
	rootCmd.RegisterFlagCompletionFunc(
		"pageSize", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"A3", "A4", "A5", "Letter", "Legal"}, cobra.ShellCompDirectiveDefault
		},
	)
	rootCmd.Flags().StringVarP(
		(*string)(&LabelPosition), "position", "p", string(pdf.LabelPosition.RIGHT),
		`Label position relative to the QR code ["T" (top)|"B" (bottom)|"L" (left)|"R" (right)]`,
	)
	rootCmd.RegisterFlagCompletionFunc(
		"position", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"T\ttop", "B\tbottom", "L\tleft", "R\tright"}, cobra.ShellCompDirectiveDefault
		},
	)
	rootCmd.Flags().StringVarP(
		&InputFileName, "inputFile", "i", "",
		"The file containing the label values. If not provided, the input will be read from stdin."+
			"Each line represents a new label. The use of a separator is optional. If there is none, label and content will be the same",
	)
	rootCmd.Flags().StringVarP(
		&Separator, "separator", "S", ";",
		`The character at which each input line should be split for the QR code content and label.`+
			`(e.g. "example.org;example text", while the QR code will be generated for example.org and "example text" will be displayed)`,
	)
	rootCmd.Flags().Float64Var(
		&PageHeight, "pageHeight", 297,
		"The page height in the specified unit. Must be set along pageWidth. The page size and the exact dimensions are mutually exclusive",
	)
	rootCmd.Flags().Float64Var(
		&PageWidth, "pageWidth", 210,
		"The page width in the specified unit. Must be set along pageHeight. The page size and the exact dimensions are mutually exclusive",
	)
	rootCmd.Flags().IntVarP(
		(*int)(&RecoveryLevel), "recoveryLevel", "R", int(qrcode.Medium),
		`The recovery level of the QR code.`+
			`There should be no need to change it unless you do experience any issues with scanning performance. [0 (low)|1 (medium)|2 (high)|3 (highest)]`,
	)
	rootCmd.RegisterFlagCompletionFunc(
		"recoveryLevel",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"0", "1", "2", "3"}, cobra.ShellCompDirectiveDefault
		},
	)
	rootCmd.MarkFlagsRequiredTogether("pageHeight", "pageWidth")
	rootCmd.MarkFlagsMutuallyExclusive("pageSize", "pageHeight")
	rootCmd.MarkFlagsMutuallyExclusive("pageSize", "pageWidth")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error executing command")
	}
}
