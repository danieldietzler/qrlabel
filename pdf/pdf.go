package pdf

import (
	"fmt"
	"github.com/go-pdf/fpdf"
	"github.com/skip2/go-qrcode"
	"io"
	"math"
	"os"
	"slices"
	"strings"
)

type PageLayout struct {
	Cell Cell
	Rows,
	Cols int
	Unit,
	SizeStr string
	Size          fpdf.SizeType
	LabelPosition Position
	FontSize      float64
}

type Cell struct {
	Width,
	Height float64
}

type Label struct {
	Content,
	Label string
}

type Position string

var LabelPosition = struct {
	TOP,
	BOTTOM,
	LEFT,
	RIGHT Position
}{
	TOP:    "T",
	BOTTOM: "B",
	LEFT:   "L",
	RIGHT:  "R",
}

func CreatePdf(
	layout PageLayout, recoveryLevel qrcode.RecoveryLevel, minQrSizePercentage float64, isBorder bool, fileName string,
	labels []Label,
) (*os.File, error) {
	var pdf *fpdf.Fpdf
	if len(layout.SizeStr) > 0 {
		pdf = fpdf.New("P", layout.Unit, layout.SizeStr, "")
	} else {
		pdf = fpdf.NewCustom(
			&fpdf.InitType{
				OrientationStr: "P",
				UnitStr:        layout.Unit,
				Size:           layout.Size,
			},
		)
	}
	pdf.SetFont("Arial", "", layout.FontSize)
	width, height := pdf.GetPageSize()
	marginWidth := math.Max(0, (width-float64(layout.Cols)*layout.Cell.Width)/2)
	marginHeight := math.Max(0, (height-float64(layout.Rows)*layout.Cell.Height)/2)
	pdf.SetLeftMargin(marginWidth)
	pdf.SetTopMargin(marginHeight)
	pdf.SetAutoPageBreak(false, marginHeight)

	pdf.AddPage()

	opt := fpdf.ImageOptions{
		ImageType:             "png",
		ReadDpi:               true,
		AllowNegativePosition: true,
	}

	for _, label := range labels {
		reader, writer := io.Pipe()

		// ensures that GetStringWidth has the right values
		pdf.SetCellMargin(0)

		imageWidth, imageHeight, cellMargin := getSizeProperties(layout, pdf, &label, minQrSizePercentage)

		pdf.SetCellMargin(cellMargin)

		generateQRCode(label.Content, recoveryLevel, writer)

		pdf.RegisterImageOptionsReader(label.Content, opt, reader)

		border := "0"
		if isBorder {
			border = "1"
		}

		x, y := pdf.GetXY()
		lines := pdf.SplitText(label.Label, layout.Cell.Width-imageWidth)

		pdf.MultiCell(layout.Cell.Width, layout.Cell.Height, "", border, "", false)

		lineCount := float64(len(lines))
		_, lineHeight := pdf.GetFontSize()
		cellWidth := layout.Cell.Width
		cellHeight := lineHeight * lineCount

		switch layout.LabelPosition {
		case LabelPosition.TOP:
			pdf.SetXY(x, y)
		case LabelPosition.BOTTOM:
			pdf.SetXY(x, y+imageHeight)
		case LabelPosition.LEFT:
			cellWidth -= imageWidth
			pdf.SetXY(x, y+(layout.Cell.Height-cellHeight)/2)
		case LabelPosition.RIGHT:
			cellWidth -= imageWidth
			pdf.SetXY(x+imageWidth, y+(layout.Cell.Height-cellHeight)/2)
		}

		pdf.MultiCell(cellWidth, lineHeight, label.Label, "", "CM", false)

		pdf.SetXY(x+layout.Cell.Width, y)

		imageXPos, imageYPos := calculateImagePosition(layout, pdf, imageWidth, imageHeight, cellMargin)

		pdf.ImageOptions(
			label.Content, imageXPos, imageYPos, imageWidth, imageHeight, false, opt, 0, "",
		)

		if pdf.GetX()+layout.Cell.Width > width-marginWidth {
			pdf.Ln(layout.Cell.Height)
		}

		if pdf.GetY()+layout.Cell.Height > height-marginHeight {
			pdf.AddPage()
		}
	}

	if !strings.HasSuffix(fileName, ".pdf") {
		fileName += ".pdf"
	}

	file, _ := os.Create(fileName)

	return file, pdf.OutputAndClose(file)
}

func generateQRCode(content string, recoveryLevel qrcode.RecoveryLevel, writer *io.PipeWriter) {
	code, _ := qrcode.New(content, recoveryLevel)
	code.DisableBorder = true

	go func() {
		code.Write(300, writer)
		writer.Close()
	}()
}

func getSizeProperties(
	layout PageLayout, pdf *fpdf.Fpdf, label *Label, minQrSizePercentage float64,
) (imageWidth, imageHeight, margin float64) {
	var minQrSize float64
	_, fontHeight := pdf.GetFontSize()
	fontSize, _ := pdf.GetFontSize()
	lines := pdf.SplitText(label.Label, layout.Cell.Width)
	smallestDimension := min(layout.Cell.Height, layout.Cell.Width)

	switch layout.LabelPosition {
	case LabelPosition.BOTTOM, LabelPosition.TOP:
		minQrSize = layout.Cell.Height * (minQrSizePercentage / 100)
		imageHeight = min(smallestDimension, layout.Cell.Height-float64(len(lines))*fontHeight)

		for imageHeight < minQrSize && fontSize > 1 {
			fontSize--
			pdf.SetFontSize(fontSize)

			lines = pdf.SplitText(label.Label, layout.Cell.Width)
			imageHeight = min(smallestDimension, layout.Cell.Height-float64(len(lines))*fontHeight)

			fmt.Fprintf(os.Stderr, "Decreased font size to %.1f to fit label\n", fontSize)
		}
		imageWidth = imageHeight

	case LabelPosition.LEFT, LabelPosition.RIGHT:
		minQrSize = layout.Cell.Width * (minQrSizePercentage / 100)
		lines = pdf.SplitText(label.Label, layout.Cell.Width-smallestDimension)
		imageWidth = smallestDimension

		for ok := true; ok; ok = fontSize > 1 {
			longestLine := slices.MaxFunc(
				lines, func(a, b string) int { return int(pdf.GetStringWidth(a)) - int(pdf.GetStringWidth(b)) },
			)

			imageWidth = max(minQrSize, layout.Cell.Width-pdf.GetStringWidth(longestLine))
			imageWidth = min(smallestDimension, imageWidth)

			if float64(len(lines))*fontHeight <= layout.Cell.Height {
				break
			}

			fontSize--
			pdf.SetFontSize(fontSize)

			lines = pdf.SplitText(label.Label, layout.Cell.Width-imageWidth)

			fmt.Fprintf(os.Stderr, "Decreased font size to %.1f to fit label\n", fontSize)
		}

		imageHeight = imageWidth
	}

	label.Label = strings.Join(lines, "\n")
	margin = min(layout.Cell.Height-imageHeight, layout.Cell.Width-imageWidth)

	return
}

func calculateImagePosition(layout PageLayout, pdf *fpdf.Fpdf, imageWidth, imageHeight, cellMargin float64) (
	imageXPos, imageYPos float64,
) {
	switch layout.LabelPosition {
	case LabelPosition.BOTTOM:
		imageXPos = pdf.GetX() - layout.Cell.Width/2 - imageHeight/2
		imageYPos = pdf.GetY() + layout.Cell.Height - imageHeight - cellMargin
	case LabelPosition.TOP:
		imageXPos = pdf.GetX() - layout.Cell.Width/2 - imageHeight/2
		imageYPos = pdf.GetY() + cellMargin
	case LabelPosition.LEFT:
		imageXPos = pdf.GetX() - imageWidth - cellMargin
		imageYPos = pdf.GetY() + layout.Cell.Height/2 - imageHeight/2
	case LabelPosition.RIGHT:
		imageXPos = pdf.GetX() - layout.Cell.Width + cellMargin
		imageYPos = pdf.GetY() + layout.Cell.Height/2 - imageHeight/2
	}

	return
}
