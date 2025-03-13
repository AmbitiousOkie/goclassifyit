package main

import (
	"embed"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed fonts/DejaVuSans-Bold.ttf
var fontData embed.FS

// BannerMode defines the banner properties: background color, text color, and text content.
type BannerMode struct {
	BgColor   color.RGBA // Background color of the banner
	TextColor color.RGBA // Text color of the banner
	Text      string     // Banner label text
}

// Predefined classification banner modes with specific colors and text labels.
var bannerModes = map[string]BannerMode{
	"cui":       {BgColor: color.RGBA{0, 255, 0, 255}, TextColor: color.RGBA{0, 0, 0, 255}, Text: "CUI"},
	"secret":    {BgColor: color.RGBA{255, 0, 0, 255}, TextColor: color.RGBA{255, 255, 255, 255}, Text: "SECRET"},
	"unclassed": {BgColor: color.RGBA{0, 0, 0, 255}, TextColor: color.RGBA{255, 255, 255, 255}, Text: "UNCLASSIFIED"},
	"custom":    {BgColor: color.RGBA{255, 255, 255, 255}, TextColor: color.RGBA{0, 0, 0, 255}, Text: "CUSTOM"},
}

func main() {
	// Define command-line flags
	dirFlag := flag.String("d", "", "Directory containing images to classify")
	fileFlag := flag.String("f", "", "Single image file to classify")
	classFlag := flag.String("c", "", "Classification type: 'unclassed', 'cui', or 'secret'")
	outputFlag := flag.String("o", "goclassifyit_output", "Output directory for classified images")
	bannerHeightFlag := flag.Int("h", 60, "Banner height in pixels (default: 60)")
	locFlag := flag.String("l", "center", "Location of banner text: 'center' (default) or 'corners'")
	textFlag := flag.String("text", "", "Custom text for banner")
	bgColorFlag := flag.String("background-color", "255,0,0", "Comma-separated R,G,B for background color (default: 255,0,0)")
	txtColorFlag := flag.String("text-color", "255,255,255", "Comma-separated R,G,B for text color (default: 255,255,255)")

	flag.Parse()

	// Validate required flags
	if *classFlag == "" {
		fmt.Println("Error: Classification type (-c) is required.")
		printUsageAndExit()
	}

	var banner BannerMode
	var exists bool

	// If classification is "custom", build a BannerMode from user-provided flags
	if *classFlag == "custom" {
		bgCol, err := parseRGB(*bgColorFlag)
		if err != nil {
			fmt.Println("Error parsing background color:", err)
			os.Exit(1)
		}
		txtCol, err := parseRGB(*txtColorFlag)
		if err != nil {
			fmt.Println("Error parsing text color:", err)
			os.Exit(1)
		}
		if *textFlag == "" {
			fmt.Println("Error: You must provide -text for custom banner mode.")
			os.Exit(1)
		}
		if *bgColorFlag == "" {
			fmt.Println("Error: You must provide -background-color for custom banner color")
		}
		if *txtColorFlag == "" {
			fmt.Println("Error: You must provide -text-color for custom banner color")
		}

		banner = BannerMode{
			BgColor:   bgCol,
			TextColor: txtCol,
			Text:      *textFlag,
		}
		exists = true // Because we created it ourselves
	} else {
		// Otherwise, look up the predefined mode
		banner, exists = bannerModes[*classFlag]
		if !exists {
			fmt.Println("Error: Invalid classification mode. Options: unclassed, cui, secret.")
			printUsageAndExit()
		}
	}

	// Validate classification mode
	if !exists {
		fmt.Println("Error: Invalid classification mode. Options: unclassed, cui, secret, custom.")
		printUsageAndExit()
	}

	if (*fileFlag == "" && *dirFlag == "") || (*fileFlag != "" && *dirFlag != "") {
		fmt.Println("Error: You must specify either a file (-f) or a directory (-d).")
		printUsageAndExit()
	}

	if *fileFlag != "" {
		if _, err := os.Stat(*fileFlag); os.IsNotExist(err) {
			fmt.Printf("Error: File '%s' does not exist.\n", *fileFlag)
			os.Exit(1)
		}

		err := processImage(*fileFlag, banner, *outputFlag, *bannerHeightFlag, *locFlag)
		if err != nil {
			fmt.Printf("Error processing file '%s': %v\n", *fileFlag, err)
			os.Exit(1)
		}
		fmt.Println("File classified successfully:", *fileFlag)
	}

	if *dirFlag != "" {
		if _, err := os.Stat(*dirFlag); os.IsNotExist(err) {
			fmt.Printf("Error: Directory '%s' does not exist.\n", *dirFlag)
			os.Exit(1)
		}

		err := processDirectory(*dirFlag, banner, *outputFlag, *bannerHeightFlag, *locFlag)
		if err != nil {
			fmt.Printf("Error processing directory '%s': %v\n", *dirFlag, err)
			os.Exit(1)
		}
		fmt.Println("All images in directory classified successfully:", *dirFlag)
	}
}

// printUsageAndExit prints usage information and exits the program.
func printUsageAndExit() {
	fmt.Println("Usage:")
	fmt.Println("  -d \"directory\"      		Classify all images in a directory")
	fmt.Println("  -f \"file\"             		Classify a specific image file")
	fmt.Println("  -c \"classification\"   		Choose classification: unclassed, cui, secret, or custom")
	fmt.Println("  -o \"output_directory\" 		Specify output directory (default: goclassifyit_output)")
	fmt.Println("  -h \"height\"          		Banner height in pixels (default: 60)")
	fmt.Println("  -l \"location\"         		Location of banner text: 'center' (default) or 'corners'")
	fmt.Println("")
	fmt.Println("When using -c custom, you must also provide:")
	fmt.Println("  -text \"some text\"      	The banner text to display")
	fmt.Println("  -background-color \"R,G,B\"  Background color (default: 255,0,0)")
	fmt.Println("  -tc \"R,G,B\"            	Text color (default: 255,255,255)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  FILE MODE:      bin/goclassifyit_linux_x64.bin -f test_images/gopher1.png -c cui -o my_output -h 80 -l corners")
	fmt.Println("  DIRECTORY MODE: bin/goclassifyit_windows_x64.exe -d test_images/ -c secret -o classified_results -h 100 -l center")
	fmt.Println("  CUSTOM MODE:	   bin/goclassifyit_linux_x64.bin -f test_images/gopher2.png -c custom -text SENSITIVE -background-color 255,255,0 -text-color 0,0,0")

	os.Exit(1)
}

func processDirectory(dirPath string, banner BannerMode, outputDir string, bannerHeight int, loc string) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var hasErrors bool // Track if any images failed

	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(dirPath, file.Name())
			err := processImage(filePath, banner, outputDir, bannerHeight, loc)
			if err != nil {
				fmt.Printf("Error processing %s: %v\n", filePath, err)
				hasErrors = true
			} else {
				fmt.Println("Classified:", filePath)
			}
		}
	}

	if hasErrors {
		return fmt.Errorf("some images failed to process")
	}
	return nil
}

// processImage loads an image, adds classification banners, and saves the result.
func processImage(imagePath string, banner BannerMode, outputDir string, bannerHeight int, loc string) error {

	// Check if the output directory is writable (simple test by creating a temp file)
	if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
		testFile := filepath.Join(outputDir, "test_write.tmp")
		f, err := os.Create(testFile)
		if err != nil {
			return fmt.Errorf("output directory '%s' is not writable: %w", outputDir, err)
		}
		f.Close()
		os.Remove(testFile)
	}

	// Open the input image file
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	// Decode the image format (supports PNG & JPEG)
	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image '%s'. Ensure the file is a valid JPEG or PNG: %w", imagePath, err)
	}

	// Validate supported formats
	if format != "jpeg" && format != "png" {
		return fmt.Errorf("unsupported image format '%s' for file: %s", format, imagePath)
	}

	// Get image dimensions
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	newHeight := height + 2*bannerHeight

	// Create a new image with extra space for banners
	newImg := image.NewRGBA(image.Rect(0, 0, width, newHeight))

	// Define banner regions
	topBanner := image.Rect(0, 0, width, bannerHeight)
	bottomBanner := image.Rect(0, newHeight-bannerHeight, width, newHeight)

	// Fill banner areas with the predefined background color
	draw.Draw(newImg, topBanner, &image.Uniform{banner.BgColor}, image.Point{}, draw.Src)
	draw.Draw(newImg, bottomBanner, &image.Uniform{banner.BgColor}, image.Point{}, draw.Src)

	// Overlay the original image onto the new image
	draw.Draw(newImg,
		image.Rect(0, bannerHeight, width, bannerHeight+height),
		img,
		bounds.Min,
		draw.Src,
	)

	// -- Load the font face once here --
	face, err := loadFontFace(36) // 36pt is an example – feel free to adjust or parameterize
	if err != nil {
		return fmt.Errorf("failed to load font face: %w", err)
	}

	// Now draw the text in either "corners" or "center" mode
	switch loc {
	case "corners":
		// 5% of width margin
		marginX := int(0.05 * float64(width))

		// Y positions for top and bottom text (vertical centering in each banner)
		topY := bannerHeight/2 + 10
		botY := (newHeight - bannerHeight/2) + 10

		// Measure the text width so we can align the right side properly
		txtWidth := measureText(face, banner.Text)

		// TOP-LEFT
		addLabel(newImg, banner.Text, marginX, topY, banner.TextColor, face)
		// TOP-RIGHT
		addLabel(newImg, banner.Text, width-marginX-txtWidth, topY, banner.TextColor, face)

		// BOTTOM-LEFT
		addLabel(newImg, banner.Text, marginX, botY, banner.TextColor, face)
		// BOTTOM-RIGHT
		addLabel(newImg, banner.Text, width-marginX-txtWidth, botY, banner.TextColor, face)

	default: // "center" or anything else
		// Center text horizontally
		topY := bannerHeight/2 + 10
		botY := (newHeight - bannerHeight/2) + 10

		// For center alignment, measure text and shift it half
		txtWidth := measureText(face, banner.Text)

		// X coordinate for center
		centerX := width/2 - (txtWidth / 2)

		// Center on top banner
		addLabel(newImg, banner.Text, centerX, topY, banner.TextColor, face)
		// Center on bottom banner
		addLabel(newImg, banner.Text, centerX, botY, banner.TextColor, face)
	}

	// Create the output directory if it does not exist
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Define the output file path
	outputPath := filepath.Join(outputDir, filepath.Base(imagePath))
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Encode and save the new image in the same format as the input
	switch format {
	case "jpeg":
		err = jpeg.Encode(outputFile, newImg, nil)
	case "png":
		err = png.Encode(outputFile, newImg)
	}

	if err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}
	return nil
}

// loadFontFace loads the embedded TTF font and returns a font.Face at a specified size.
func loadFontFace(fontSize float64) (font.Face, error) {
	fontBytes, err := fontData.ReadFile("fonts/DejaVuSans-Bold.ttf")
	if err != nil {
		return nil, fmt.Errorf("unable to read embedded font: %w", err)
	}
	tt, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse font: %w", err)
	}
	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create font face: %w", err)
	}
	return face, nil
}

// measureText returns the width of the given text (in pixels) for the specified font face.
func measureText(face font.Face, text string) int {
	d := &font.Drawer{
		Face: face,
	}
	return d.MeasureString(text).Round()
}

// addLabel draws the text at the exact (x, y) coordinate without additional horizontal shifting.
func addLabel(img *image.RGBA, text string, x, y int, col color.RGBA, face font.Face) {
	d := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{col},
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.I(x),
			Y: fixed.I(y),
		},
	}
	d.DrawString(text)
}

func parseRGB(str string) (color.RGBA, error) {
	var r, g, b int
	_, err := fmt.Sscanf(str, "%d,%d,%d", &r, &g, &b)
	if err != nil {
		return color.RGBA{}, fmt.Errorf("invalid color format (expected \"R,G,B\"): %w", err)
	}
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
		return color.RGBA{}, fmt.Errorf("invalid color value, each must be between 0 and 255")
	}
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}, nil
}
