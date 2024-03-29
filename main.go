package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"os"
	"strings"

	"github.com/muni-corn/brite"
)

func main() {
	// set variables from arguments, if any
	var imageFileName, templateFileName, outputFileName string
	for k, v := range os.Args {
		if k+1 >= len(os.Args) {
			break
		}
		next := os.Args[k+1]
		switch v {
		case "-l", "--light-color":
			lightColor.primary = next
			lightColor.secondary = next + "c0"
		case "-d", "--dark-color":
			darkColor.primary = next
			darkColor.secondary = next + "c0"
		case "-i", "--image-file", "-w", "--wallpaper-file":
			imageFileName = next
		case "-t", "--template-file":
			templateFileName = next
		case "-o", "--output-file":
			outputFileName = next
		}
	}

	imageFileMissing := imageFileName == ""
	templateFileMissing := templateFileName == ""
	outputFileMissing := outputFileName == ""
	if imageFileMissing || templateFileMissing || outputFileMissing {
		fmt.Println("okay, i don't have these files to do my job:")
		if imageFileMissing {
			fmt.Println("    - an image file")
		}
		if templateFileMissing {
			fmt.Println("    - a template file")
		}
		if outputFileMissing {
			fmt.Println("    - an output file")
		}

		displayHelp()
		return
	}

	// open image file
	imageFile, err := os.Open(imageFileName)
	if err != nil {
		panic(err)
	}
	defer imageFile.Close()

	// open template file
	templateFile, err := os.Open(templateFileName)
	if err != nil {
		panic(err)
	}
	defer templateFile.Close()

	// open output file
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	// decode wallpaper
	img, _, err := image.Decode(imageFile)
	if err != nil {
		panic(err)
	}

    // dimensions and stuff
	quarterWidth := img.Bounds().Dx() / 4
    topOfImage := getTopOfImage(img)
    fmt.Println("top of image:", topOfImage)

	// get brightness classification
	leftStatusBounds := image.Rect(0, topOfImage, quarterWidth, topOfImage + 32)
	rightStatusBounds := image.Rect(quarterWidth*3, topOfImage, quarterWidth*4, topOfImage + 32)

	// get brightness of the left and right status bar
	// sections
	leftBrightness, _ := brite.GetImageBrightnessBounds(img, leftStatusBounds)
	fmt.Printf("left brightness: %s\n", string(leftBrightness))

	rightBrightness, _ := brite.GetImageBrightnessBounds(img, rightStatusBounds)
	fmt.Printf("right brightness: %s\n", string(rightBrightness))

	// assign color pairs
	var leftColorPair, rightColorPair colorPair
	leftColorPair = assignColorPair(leftBrightness)
	rightColorPair = assignColorPair(rightBrightness)

	// see: function name
	replaceTemplateStrings(templateFile, outputFile, leftColorPair, rightColorPair)
}

const screenAspectRatio = 16.0 / 9.0

// gets the top of the visible wallpaper
func getTopOfImage(img image.Image) int {
    width, height := img.Bounds().Dx(), img.Bounds().Dy()

    visibleHeight := float32(width)/screenAspectRatio
    top := float32(height/2) - visibleHeight/2

    return int(top)
}

// returns a respective colorPair based on the brightness of
// the section
func assignColorPair(sectionBrightness brite.ImageBrightness) colorPair {
	switch sectionBrightness {
	case brite.Dark:
		return lightColor
	case brite.Light:
		return darkColor
	}

	return darkColor
}

func replaceTemplateStrings(templateFile, outputFile *os.File, left colorPair, right colorPair) {
	all, err := ioutil.ReadAll(templateFile)
	if err != nil {
		panic(err)
	}

	fileContentsString := string(all)

	// fill in left colors
	fileContentsString = strings.Replace(fileContentsString, leftPrimaryTemplateString, left.primary, -1)
	fileContentsString = strings.Replace(fileContentsString, leftSecondaryTemplateString, left.secondary, -1)

	// fill in right colors
	fileContentsString = strings.Replace(fileContentsString, rightPrimaryTemplateString, right.primary, -1)
	fileContentsString = strings.Replace(fileContentsString, rightSecondaryTemplateString, right.secondary, -1)

	outputFile.WriteString(fileContentsString)
}

func displayHelp() {
	fmt.Printf(`
do something like this:"

    status-bar-template-parsing-thing -i "/path/to/wallpaper.jpg" -t "/path/to/template.config" -o "/path/to/output.config"

options:

    -i, --image-file, -w, --wallpaper-file
        path to a wallpaper you want to use

    -t, --template-file
        path to a template you want to use
        put these in your template and they will be replaced with colors (e.g. rrggbbaa)
        "%s" should be used for bold colors on the left
        "%s" should be used for bold colors on the right
        "%s" should be used for faded colors on the left
        "%s" should be used for faded colors on the right

    -o, --output-file
        path to the output file

    -l, --light-color
        lighter color you want to use

    -d, --dark-color
        darker color you want to use` + "\n", leftPrimaryTemplateString, rightPrimaryTemplateString, leftSecondaryTemplateString, rightSecondaryTemplateString)

}

type colorPair struct {
	primary   string
	secondary string
}

var (
	lightColor = colorPair{
		primary:   "ffffff",
		secondary: "ffffff80",
	}

	darkColor = colorPair{
		primary:   "000000",
		secondary: "00000080",
	}
)

const (
	leftPrimaryTemplateString    = "#L_PRI#"
	rightPrimaryTemplateString   = "#R_PRI#"
	leftSecondaryTemplateString  = "#L_SEC#"
	rightSecondaryTemplateString = "#R_SEC#"
)
