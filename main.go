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
	var imageFileName, templateFileName string
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
		}
	}

	if imageFileName == "" && templateFileName == "" {
		fmt.Println("what the heck???? you didn't specify an image file OR a template file")
		displayHelp()
		return
	} else if imageFileName == "" {
		fmt.Println("hey there friend, specify me an image file")
		displayHelp()
		return
	} else if templateFileName == "" {
		fmt.Println("i need a template file to work")
		displayHelp()
		return
	}

	// open image file
	imageFile, err := os.Open(imageFileName)
	if err != nil {
		panic(err)
	}

	// open template file
	templateFile, err := os.Open(templateFileName)
	if err != nil {
		panic(err)
	}

	// decode wallpaper
	img, _, err := image.Decode(imageFile)
	if err != nil {
		panic(err)
	}

	// get brightness classification
	quarterWidth := img.Bounds().Dx() / 4
	leftStatusBounds := image.Rect(0, 0, quarterWidth, 32)
	rightStatusBounds := image.Rect(quarterWidth*3, 0, quarterWidth, 32)

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
	replaceTemplateStrings(templateFile, leftColorPair, rightColorPair)
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

func replaceTemplateStrings(templateFile *os.File, left colorPair, right colorPair) {
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

	// close the file. we'll reopen it for writing back to
	// it
	err = templateFile.Close()
	if err != nil {
		// THERE ARE SO MANY FREAKING PANIC CALLS I'M
		// STRESSING OUT
		panic(err)
	}

	// re-open the file to write to it
	templateFile, err = os.Create(templateFile.Name())
	if err != nil {
		panic(err)
	}
	templateFile.WriteString(fileContentsString)
	defer templateFile.Close()
}

func displayHelp() {
	fmt.Println(`
do something like this:"

    status-bar-template-parsing-thing -i /path/to/wallpaper.jpg -t /path/to/template.config

options:

    -i, --image-file, -w, --wallpaper-file
        path to a wallpaper you want to use

    -t, --template-file"
        path to a template you want to use
        put these in your template and they will be replaced with colors (e.g. rrggbbaa)
        " + leftPrimaryTemplateString + " should be used for bold colors on the left
        " + rightPrimaryTemplateString + " should be used for bold colors on the right
        " + leftSecondaryTemplateString + " should be used for faded colors on the left
        " + rightSecondaryTemplateString + " should be used for faded colors on the right

    -l, --light-color
        lighter color you want to use

    -d, --dark-color
        darker color you want to use`)

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
