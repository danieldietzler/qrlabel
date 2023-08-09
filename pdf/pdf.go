package pdf

import (
	"fmt"
	"github.com/go-pdf/fpdf"
	"github.com/skip2/go-qrcode"
	"io"
	"math"
	"os"
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

func CreatePdf(layout PageLayout, recoveryLevel qrcode.RecoveryLevel, fileName string, labels []Label) (
	*os.File, error,
) {
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
	marginWidth := (width - float64(layout.Cols)*layout.Cell.Width) / 2
	marginHeight := (height - float64(layout.Rows)*layout.Cell.Height) / 2
	pdf.SetLeftMargin(marginWidth)
	pdf.SetTopMargin(marginHeight)
	pdf.SetAutoPageBreak(true, marginHeight)

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

		imageWidth, imageHeight, cellMargin, alignString := getSizeProperties(layout, pdf, label)

		pdf.SetCellMargin(cellMargin)

		generateQRCode(label.Content, recoveryLevel, writer)

		pdf.RegisterImageOptionsReader(label.Content, opt, reader)

		pdf.CellFormat(
			layout.Cell.Width, layout.Cell.Height, fmt.Sprintf(label.Label), "1", 0, alignString, false, 0, "",
		)

		imageXPos, imageYPos := calculateImagePosition(layout, pdf, imageWidth, imageHeight, cellMargin)

		pdf.ImageOptions(
			label.Content, imageXPos, imageYPos, imageWidth, imageHeight, false, opt, 0, "",
		)

		if pdf.GetX()+cellMargin > width-marginWidth {
			pdf.Ln(-1)
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

func getSizeProperties(layout PageLayout, pdf *fpdf.Fpdf, label Label) (
	imageWidth, imageHeight, margin float64, alignString string,
) {
	_, fontHeight := pdf.GetFontSize()

	imageHeight = math.Min(layout.Cell.Height, math.Min(layout.Cell.Width, layout.Cell.Height-fontHeight))
	imageWidth = math.Min(
		layout.Cell.Height, math.Min(layout.Cell.Width, layout.Cell.Width-pdf.GetStringWidth(label.Label)),
	)

	switch layout.LabelPosition {
	case LabelPosition.BOTTOM:
		margin = (layout.Cell.Height - (imageHeight + fontHeight)) / 3
		alignString = "CB"
		imageWidth = 0
	case LabelPosition.TOP:
		margin = (layout.Cell.Height - (imageHeight + fontHeight)) / 3
		alignString = "CT"
		imageWidth = 0
	case LabelPosition.LEFT:
		margin = (layout.Cell.Width - (imageWidth + pdf.GetStringWidth(label.Label))) / 3
		alignString = "LM"
		imageHeight = 0
	case LabelPosition.RIGHT:
		margin = (layout.Cell.Width - (imageWidth + pdf.GetStringWidth(label.Label))) / 3
		alignString = "RM"
		imageHeight = 0
	}

	return
}

func calculateImagePosition(layout PageLayout, pdf *fpdf.Fpdf, imageWidth, imageHeight, cellMargin float64) (
	imageXPos, imageYPos float64,
) {
	switch layout.LabelPosition {
	case LabelPosition.BOTTOM:
		imageXPos = pdf.GetX() - layout.Cell.Width/2 - imageHeight/2
		imageYPos = pdf.GetY() + cellMargin
	case LabelPosition.TOP:
		imageXPos = pdf.GetX() - layout.Cell.Width/2 - imageHeight/2
		imageYPos = pdf.GetY() + layout.Cell.Height - imageHeight - cellMargin
	case LabelPosition.LEFT:
		imageXPos = pdf.GetX() - imageWidth - cellMargin
		imageYPos = pdf.GetY()
	case LabelPosition.RIGHT:
		imageXPos = pdf.GetX() - layout.Cell.Width + cellMargin
		imageYPos = pdf.GetY()
	}

	return
}
