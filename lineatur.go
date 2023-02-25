package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

// https://de.wikipedia.org/wiki/Lineatur
// Winkel ist von der Grundlinie aus zur Schräge nach oben gemessen
//    1:1:1 Sütterlinschrift (1915 - 1941)
//    2:3:2 Offenbacher Schrift (1927) (75°-80°)
//    3:4:3 Offenbacher Schrift, Lateinische Ausgangsschrift
//    2:1:2 Deutsche Kurrentschrift (60°)
//    3:2:3 Copperplate (Winkel: 52°-60°)

func usage() {
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "Line proportions: no argument = just one line\n")
	fmt.Fprintf(os.Stderr, "Line proportions: num = two lines (the value doesn't matter)\n")
	fmt.Fprintf(os.Stderr, "Line proportions: num[:num...]\n")
	fmt.Fprintf(os.Stderr, "Slanted helper lines: \"num:num\" the angle and number per line of slanted helper lines\n")
	fmt.Fprintf(os.Stderr, "Page margins: num:num:num:num top, right, bottom and left margins of the page in mm\n")
	fmt.Fprintf(os.Stderr, "examples:\n")
	fmt.Fprintf(os.Stderr, "    -p 2:1:2 -s 60:10  Deutsche Kurrentschrift\n")
	fmt.Fprintf(os.Stderr, "    -p 1:1:1           Sütterlinschrift\n")
	fmt.Fprintf(os.Stderr, "    -p 2:3:2 -s 75:10  Offenbacher Schrift\n")
	fmt.Fprintf(os.Stderr, "    -p 3:4:3           Offenbacher Schrift, Lateinische Ausgangsschrift\n")
	fmt.Fprintf(os.Stderr, "    -p 3:2:3 -s 52:10  Copperplate\n")
}

type PaperSize struct {
	Width  float64 // mm
	Height float64 // mm
}

/*
	size explanation: https://unsharpen.com/paper-sizes/
*/
var PaperSizes = map[string]PaperSize{
	"A5":      PaperSize{148.0, 210.0},
	"A4":      PaperSize{210.0, 297.0},
	"Invoice": PaperSize{140.0, 216.0},
	"Legal":   PaperSize{203.0, 330.0},
	"Letter":  PaperSize{216.0, 279.0},
}

func parseMultiUint64(s string) ([]float64, error) {
	if s == "" {
		return nil, nil
	}
	strs := strings.Split(s, ":")
	values := []float64{}
	for _, m := range strs {
		u, err := strconv.ParseUint(m, 10, 64)
		if err != nil {
			return nil, err
		}
		values = append(values, float64(u))
	}
	return values, nil
}

func drawLineatur(pdf *gofpdf.Fpdf, x, y, lineHeight, width float64, lineDists []float64, lineWidth float64, slants []float64) {
	pdf.SetLineWidth(lineWidth)
	switch len(lineDists) {
	case 0:
		pdf.MoveTo(x, y+lineHeight)
		pdf.LineTo(x+width, y+lineHeight)
		pdf.DrawPath("D")
	default:
		_y := y
		pdf.MoveTo(x, _y)
		pdf.LineTo(x+width, _y)
		pdf.DrawPath("D")
		for _, d := range lineDists {
			_y += d
			pdf.MoveTo(x, _y)
			pdf.LineTo(x+width, _y)
			pdf.DrawPath("D")
		}
		// draw lines left and right
		pdf.MoveTo(x, y)
		pdf.LineTo(x, y+lineHeight)
		pdf.DrawPath("D")
		pdf.MoveTo(x+width, y)
		pdf.LineTo(x+width, y+lineHeight)
		pdf.DrawPath("D")
	}
	// draw slanted helper lines
	if len(slants) == 2 {
		angle := math.Pi * (90.0 - slants[0]) / 180.0
		b := math.Abs(lineHeight * math.Tan(angle))
		n := (width - b) / (slants[1] - 1)
		for i := 0.0; i < slants[1]; i++ {
			_x := x + n*i
			if slants[0] <= 90 {
				pdf.MoveTo(_x, y+lineHeight)
				pdf.LineTo(_x+b, y)
			} else {
				pdf.MoveTo(_x+b, y+lineHeight)
				pdf.LineTo(_x, y)
			}
			pdf.DrawPath("D")
		}
	}
}

func proportionsToLengths(proportions []float64, lineHeight float64) []float64 {
	lineDists := []float64{}
	// sum of proportions
	sumProp := 0.0
	for _, p := range proportions {
		sumProp += p
	}
	// absolute lengths for proportions
	for _, p := range proportions {
		lineDists = append(lineDists, lineHeight*p/sumProp)
	}
	return lineDists
}

func drawAllLineatur(pdf *gofpdf.Fpdf, paperSize PaperSize, margins []float64, lineHeight float64, lineSpacing float64, proportions []float64, slants []float64, lineWidth float64) {
	lineDists := proportionsToLengths(proportions, lineHeight)
	width := paperSize.Width - margins[1] - margins[3]
	x := margins[3]
	y := margins[0]
	for (y + lineHeight) < (paperSize.Height - margins[2]) {
		drawLineatur(pdf, x, y, lineHeight, width, lineDists, lineWidth, slants)
		y += lineHeight + lineSpacing
	}
}

func main() {
	var paperSize, _proportions, _slants, _margins, filename string
	var lineHeight, lineSpacing uint64
	var lineWidth float64
	flag.StringVar(&filename, "o", "output.pdf", "output file")
	flag.StringVar(&paperSize, "ps", "A4", "Paper size of your printer. Possible values: A5, A4, Invoice, Legal, Letter. Print without scaling.")
	flag.StringVar(&_proportions, "p", "", "Line proportions.")
	flag.StringVar(&_slants, "s", "", "Slanted helper lines.")
	flag.StringVar(&_margins, "m", "5:15:15:5", "Page margins.")
	flag.Uint64Var(&lineHeight, "lh", 10, "Line height in mm.")
	flag.Uint64Var(&lineSpacing, "ls", 5, "Line spacing in mm.")
	flag.Float64Var(&lineWidth, "lw", 0.3, "Line width in mm.")
	flag.Usage = usage
	flag.Parse()
	if _, ok := PaperSizes[paperSize]; !ok {
		fmt.Printf("paper size \"%s\" choosen for printing is unknown/not allowed\n", paperSize)
		os.Exit(1)
	}
	proportions, err := parseMultiUint64(_proportions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wrong arguments for -p: %s\n", _proportions)
		os.Exit(1)
	}
	slants, err := parseMultiUint64(_slants)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wrong arguments for -s: %s\n", _slants)
		os.Exit(1)
	}
	if len(slants) != 0 && len(slants) != 2 {
		fmt.Fprintf(os.Stderr, "wrong number of arguments for -s: %s\n", _slants)
		os.Exit(1)
	}
	/*
		if len(slants) == 2 && (slants[0] > 90) {
			fmt.Fprintf(os.Stderr, "value out of interval for parameter -s: %s\n", _slants)
			os.Exit(1)
		}
	*/
	margins, err := parseMultiUint64(_margins)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wrong arguments for -m: %s\n", _margins)
		os.Exit(1)
	}
	if len(margins) != 0 && len(margins) != 4 {
		fmt.Fprintf(os.Stderr, "wrong number of arguments for -m: %s\n", _margins)
		os.Exit(1)
	}

	// Initialize the graphic context on a pdf document
	pdf := gofpdf.New("P", "mm", paperSize, "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()
	drawAllLineatur(pdf, PaperSizes[paperSize], margins, float64(lineHeight), float64(lineSpacing), proportions, slants, lineWidth)
	pdf.OutputFileAndClose(filename)
}
