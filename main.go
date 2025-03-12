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
}

func main() {
	// Define command-line flags
	dirFlag := flag.String("d", "", "Directory containing images to classify")
	fileFlag := flag.String("f", "", "Single image file to classify")
	classFlag := flag.String("c", "", "Classification type: 'unclassed', 'cui', or 'secret'")
	outputFlag := flag.String("o", "goclassy_output", "Output directory for classified images")
	bannerHeightFlag := flag.Int("h", 60, "Banner height in pixels (default: 60)")

	flag.Parse()

	// Validate required flags
	if *classFlag == "" {
		fmt.Println("Error: Classification type (-c) is required.")
		printUsageAndExit()
	}

	if (*fileFlag == "" && *dirFlag == "") || (*fileFlag != "" && *dirFlag != "") {
		fmt.Println("Error: You must specify either a file (-f) or a directory (-d).")
		printUsageAndExit()
	}

	// Validate classification mode
	banner, exists := bannerModes[*classFlag]
	if !exists {
		fmt.Println("Error: Invalid classification mode. Available options: unclassed, cui, secret.")
		printUsageAndExit()
	}

	if *fileFlag != "" {
		if _, err := os.Stat(*fileFlag); os.IsNotExist(err) {
			fmt.Printf("Error: File '%s' does not exist.\n", *fileFlag)
			os.Exit(1)
		}

		err := processImage(*fileFlag, banner, *outputFlag, *bannerHeightFlag)
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

		err := processDirectory(*dirFlag, banner, *outputFlag, *bannerHeightFlag)
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
	fmt.Println("  -d \"directory\"        Classify all images in a directory")
	fmt.Println("  -f \"file\"             Classify a specific image file")
	fmt.Println("  -c \"classification\"   Choose classification: unclassed, cui, or secret")
	fmt.Println("  -o \"output_directory\" Specify output directory (default: goclassy_output)")
	fmt.Println("  -h \"height\"           Banner height in pixels (default: 60)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  FILE MODE: 		goclass -f test.png -c cui -o my_output -h 80")
	fmt.Println("  DIRECTORY MODE:	goclass -d images/ -c secret -o classified_results -h 100")
	os.Exit(1)
}

func processDirectory(dirPath string, banner BannerMode, outputDir string, bannerHeight int) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var hasErrors bool // Track if any images failed

	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(dirPath, file.Name())
			err := processImage(filePath, banner, outputDir, bannerHeight)
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
func processImage(imagePath string, banner BannerMode, outputDir string, bannerHeight int) error {

	// Check if the output directory is writable
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
		return fmt.Errorf("unsupported image format '%s' for file: %s. Only JPEG and PNG are supported", format, imagePath)
	}

	// Get image dimensions
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	newHeight := height + (2 * bannerHeight)

	// Create a new image with extra space for banners
	newImg := image.NewRGBA(image.Rect(0, 0, width, newHeight))

	// Define banner regions
	topBanner := image.Rect(0, 0, width, bannerHeight)
	bottomBanner := image.Rect(0, newHeight-bannerHeight, width, newHeight)

	// Fill banner areas with the predefined background color
	draw.Draw(newImg, topBanner, &image.Uniform{banner.BgColor}, image.Point{}, draw.Src)
	draw.Draw(newImg, bottomBanner, &image.Uniform{banner.BgColor}, image.Point{}, draw.Src)

	// Overlay the original image onto the new image, centering it between the banners
	draw.Draw(newImg, image.Rect(0, bannerHeight, width, bannerHeight+height), img, bounds.Min, draw.Src)

	// Add classification text to banners
	addLabel(newImg, banner.Text, width/2, bannerHeight/2+10, banner.TextColor)           // Top banner
	addLabel(newImg, banner.Text, width/2, newHeight-bannerHeight/2+10, banner.TextColor) // Bottom banner

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
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	return nil
}

func addLabel(img *image.RGBA, text string, x, y int, color color.RGBA) error {
	// Load the embedded font
	fontBytes, err := fontData.ReadFile("fonts/DejaVuSans-Bold.ttf")
	if err != nil {
		return fmt.Errorf("failed to load embedded font: %w", err)
	}

	// Parse the font
	f, err := opentype.Parse(fontBytes)
	if err != nil {
		return fmt.Errorf("failed to parse font: %w", err)
	}

	// Adjust font size dynamically based on banner height
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    36,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return fmt.Errorf("failed to create font face: %w", err)
	}

	// Position text in the center
	point := fixed.Point26_6{X: fixed.I(x - len(text)*8), Y: fixed.I(y)}

	// Draw text onto the image
	d := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{color},
		Face: face,
		Dot:  point,
	}
	d.DrawString(text)
	return nil
}
