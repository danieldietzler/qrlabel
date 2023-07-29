package pdf

import (
	"fmt"
	"github.com/go-pdf/fpdf"
	"github.com/skip2/go-qrcode"
	"io"
	"math"
	"os"
)

type PageLayout struct {
	Cell Cell
	Rows,
	Cols float64
	Unit,
	SizeStr string
	LabelOrientation Orientation
}

type Cell struct {
	Width,
	Height float64
}

type Label struct {
	Content,
	Label string
}

type Orientation string

var LabelOrientation = struct {
	TOP,
	BOTTOM,
	LEFT,
	RIGHT Orientation
}{
	TOP:    "T",
	BOTTOM: "B",
	LEFT:   "L",
	RIGHT:  "R",
}

func CreatePdf(layout PageLayout, fileName string, labels []Label) *os.File {
	reader, writer := io.Pipe()
	pdf := fpdf.New("P", layout.Unit, layout.SizeStr, "")
	pdf.SetFont("Arial", "", 12)
	width, height := pdf.GetPageSize()
	marginWidth := (width - layout.Cols*layout.Cell.Width) / 2
	marginHeight := (height - layout.Rows*layout.Cell.Height) / 2
	pdf.SetLeftMargin(marginWidth)
	pdf.SetTopMargin(marginHeight)
	pdf.SetAutoPageBreak(true, marginHeight)

	pdf.AddPage()

	for _, label := range labels {
		pdf.SetCellMargin(0)

		opt := fpdf.ImageOptions{
			ImageType:             "png",
			ReadDpi:               true,
			AllowNegativePosition: true,
		}

		generateQRCode(label.Content, writer)

		pdf.RegisterImageOptionsReader("code", opt, reader)

		imageXPos, imageYPos, margin, imageWidth, imageHeight, alignString := calculatePositions(layout, pdf, label)

		pdf.SetCellMargin(margin)

		pdf.CellFormat(
			layout.Cell.Width, layout.Cell.Height, fmt.Sprintf(label.Label), "1", 0, alignString, false, 0, "",
		)

		pdf.ImageOptions(
			"code", imageXPos, imageYPos, imageWidth, imageHeight, false, opt, 0, "",
		)

		if pdf.GetX()+layout.Cell.Width > width-marginWidth {
			pdf.Ln(-1)
		}
	}

	file, _ := os.Create(fileName)
	pdf.OutputAndClose(file)
	return file
}

func generateQRCode(content string, writer *io.PipeWriter) {
	code, _ := qrcode.New(content, qrcode.Medium)
	code.DisableBorder = true

	go func() {
		code.Write(300, writer)
		defer writer.Close()
	}()
}

func calculatePositions(layout PageLayout, pdf *fpdf.Fpdf, label Label) (
	imageXPos, imageYPos, margin, imageWidth, imageHeight float64, alignString string,
) {
	_, fontHeight := pdf.GetFontSize()

	imageHeight = math.Min(layout.Cell.Height, math.Min(layout.Cell.Width, layout.Cell.Height-fontHeight))
	imageWidth = math.Min(
		layout.Cell.Height, math.Min(layout.Cell.Width, layout.Cell.Width-pdf.GetStringWidth(label.Label)),
	)

	switch layout.LabelOrientation {
	case LabelOrientation.BOTTOM:
		margin = (layout.Cell.Height - (imageHeight + fontHeight)) / 3
		imageXPos = pdf.GetX() + layout.Cell.Width/2 - imageHeight/2
		imageYPos = pdf.GetY() + margin
		alignString = "CB"
		imageWidth = 0
	case LabelOrientation.TOP:
		margin = (layout.Cell.Height - (imageHeight + fontHeight)) / 3
		imageXPos = pdf.GetX() + layout.Cell.Width/2 - imageHeight/2
		imageYPos = pdf.GetY() + layout.Cell.Height - imageHeight - margin
		alignString = "CT"
		imageWidth = 0
	case LabelOrientation.LEFT:
		margin = (layout.Cell.Width - (imageWidth + pdf.GetStringWidth(label.Label))) / 3
		imageXPos = pdf.GetX() + layout.Cell.Width - imageWidth - margin
		imageYPos = pdf.GetY()
		alignString = "LM"
		imageHeight = 0
	case LabelOrientation.RIGHT:
		margin = (layout.Cell.Width - (imageWidth + pdf.GetStringWidth(label.Label))) / 3
		imageXPos = pdf.GetX() + margin
		imageYPos = pdf.GetY()
		alignString = "RM"
		imageHeight = 0
	}

	return
}
