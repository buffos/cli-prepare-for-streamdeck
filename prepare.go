package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/buffos/cli-prepare-for-streamdeck/config"
	"github.com/disintegration/imaging"
)

const THUMB_WIDTH = 144

type OscCommand struct {
	OscPath  string `json:"osc_path"`
	OscValue []any  `json:"osc_value"`
	OscPort  int    `json:"osc_port"`
}

type MediaEntry struct {
	Title        string       `json:"title"`
	Image        string       `json:"image"`
	ImagePressed string       `json:"image_pressed"`
	OscCommands  []OscCommand `json:"osc_commands"`
	FullPath     string       `json:"full_path"`
	Scripts      []string     `json:"scripts"`
	ScriptPaths  []string     `json:"script_paths"`
	Delays       []int        `json:"delays"`
}

type MediaConfig struct {
	Files   []MediaEntry `json:"files"`
	OscRoot string       `json:"osc_root_path"`
	OscArg  []string     `json:"osc_arg"`
}

// hexToRGBA converts a hex color string (#RRGGBB) to color.RGBA
func hexToRGBA(hex string) (color.RGBA, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return color.RGBA{}, fmt.Errorf("invalid hex color length")
	}

	var r, g, b uint8
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return color.RGBA{r, g, b, 255}, nil
}

// addBorder adds a colored border to an image
func addBorder(img image.Image, borderWidth int, borderColor color.RGBA) *image.NRGBA {
	bounds := img.Bounds()
	newWidth := bounds.Dx() + (borderWidth * 2)
	newHeight := bounds.Dy() + (borderWidth * 2)

	bordered := imaging.New(newWidth, newHeight, borderColor)

	// Draw the original image in the center
	bordered = imaging.Paste(bordered, img, image.Point{borderWidth, borderWidth})

	return bordered
}

// createResizedImage creates a thumbnail with specified width maintaining aspect ratio
func createResizedImage(sourcePath, targetPath string, width int) error {
	// Open the source image
	img, err := imaging.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %v", err)
	}

	// Calculate new dimensions maintaining aspect ratio
	bounds := img.Bounds()
	ratio := float64(width) / float64(bounds.Dx())
	height := int(float64(bounds.Dy()) * ratio)

	// Resize the image
	resized := imaging.Resize(img, width, height, imaging.Lanczos)

	// Save the resized image
	err = imaging.Save(resized, targetPath)
	if err != nil {
		return fmt.Errorf("failed to save resized image: %v", err)
	}

	return nil
}

func processMediaFiles(searchPath string, mediaType config.MediaType, oscOption config.OscPrefixOption, borderColorHex string, borderWidth int) error {
	var entries []MediaEntry
	var validExtensions []string
	var fullPaths []string

	// Convert hex color to RGBA
	borderColor, err := hexToRGBA(borderColorHex)
	if err != nil {
		return fmt.Errorf("invalid border color: %v", err)
	}

	switch mediaType {
	case config.ImageType:
		validExtensions = []string{".jpg", ".jpeg", ".png", ".gif"}
	case config.VideoType:
		validExtensions = []string{".mp4", ".avi", ".mov", ".mkv"}
	case config.AudioType:
		validExtensions = []string{".mp3", ".wav", ".ogg", ".flac"}
	}

	err = filepath.Walk(searchPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != searchPath {
			return filepath.SkipDir
		}

		ext := strings.ToLower(filepath.Ext(path))
		for _, validExt := range validExtensions {
			if ext == validExt {
				entry := processFile(path, mediaType, len(entries), oscOption, borderColor, borderWidth)
				entries = append(entries, entry)
				fullPaths = append(fullPaths, path)
				break
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking through directory: %v", err)
	}

	config := MediaConfig{
		Files:   entries,
		OscRoot: "",
		OscArg:  fullPaths,
	}

	jsonData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("error creating JSON: %v", err)
	}

	jsonPath := filepath.Join(searchPath, "media_config.json")
	err = os.WriteFile(jsonPath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error saving JSON file: %v", err)
	}

	fmt.Printf("Successfully processed %d files. Configuration saved to %s\n", len(entries), jsonPath)
	return nil
}

func processFile(filePath string, mediaType config.MediaType, index int, oscOption config.OscPrefixOption, borderColor color.RGBA, borderWidth int) MediaEntry { //nosonar
	fileName := filepath.Base(filePath)
	fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	ext := filepath.Ext(fileName)

	// Ensure 2-digit index
	indexStr := fmt.Sprintf("%02d", index+1)

	var oscPath string
	var oscValue int

	if oscOption.AugmentIndex {
		oscPath = fmt.Sprintf("%s%s", oscOption.Prefix, indexStr)
	} else {
		oscPath = oscOption.Prefix
	}

	if oscOption.ArgumentType == "serial" {
		oscValue = oscOption.ArgumentBase + index
	} else if oscOption.ArgumentType == "constant" {
		oscValue = oscOption.ArgumentBase
	}

	oscCommand := OscCommand{
		OscPath:  oscPath,
		OscValue: []any{oscValue},
		OscPort:  8000,
	}

	entry := MediaEntry{
		Title:       fileNameWithoutExt,
		OscCommands: []OscCommand{oscCommand},
		FullPath:    filePath,
		Scripts:     []string{},
		ScriptPaths: []string{},
		Delays:      []int{},
	}

	switch mediaType {
	case config.ImageType:
		// Create thumbnail
		thumbName := fileNameWithoutExt + "_thumb" + ext
		thumbPath := filepath.Join(filepath.Dir(filePath), thumbName)

		if err := createResizedImage(filePath, thumbPath, THUMB_WIDTH); err != nil {
			fmt.Printf("Error creating thumbnail for %s: %v\n", fileName, err)
			return entry
		}

		entry.Image = thumbName

		// Create pressed version from thumbnail
		pressedName := fileNameWithoutExt + "_pressed" + ext
		pressedPath := filepath.Join(filepath.Dir(filePath), pressedName)

		if err := createPressedImage(thumbPath, pressedPath, borderColor, borderWidth); err != nil {
			fmt.Printf("Error creating pressed image for %s: %v\n", fileName, err)
		} else {
			entry.ImagePressed = pressedName
		}

	case config.VideoType:
		// Create thumbnail from first frame
		thumbName := fileNameWithoutExt + "_thumb.jpg"
		thumbPath := filepath.Join(filepath.Dir(filePath), thumbName)

		if err := extractVideoThumbnail(filePath, thumbPath); err != nil {
			fmt.Printf("Error extracting thumbnail for %s: %v\n", fileName, err)
			return entry
		}

		entry.Image = thumbName

		// Create pressed version from thumbnail
		pressedName := fileNameWithoutExt + "_pressed.jpg"
		pressedPath := filepath.Join(filepath.Dir(filePath), pressedName)

		if err := createPressedImage(thumbPath, pressedPath, borderColor, borderWidth); err != nil {
			fmt.Printf("Error creating pressed thumbnail for %s: %v\n", fileName, err)
		} else {
			entry.ImagePressed = pressedName
		}

	case config.AudioType:
		entry.Image = ""
		entry.ImagePressed = ""
	}

	return entry
}

func createPressedImage(sourcePath, targetPath string, borderColor color.RGBA, borderWidth int) error {
	// Open the source image
	img, err := imaging.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %v", err)
	}

	// Add border
	bordered := addBorder(img, borderWidth, borderColor)

	// Save the bordered image
	err = imaging.Save(bordered, targetPath)
	if err != nil {
		return fmt.Errorf("failed to save bordered image: %v", err)
	}

	return nil
}

func extractVideoThumbnail(videoPath, thumbnailPath string) error {
	// Extract first frame using ffmpeg
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vframes", "1", "-f", "image2", thumbnailPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract frame: %v", err)
	}

	// Resize the extracted frame to thumbnail size
	if err := createResizedImage(thumbnailPath, thumbnailPath, THUMB_WIDTH); err != nil {
		return fmt.Errorf("failed to resize video thumbnail: %v", err)
	}

	return nil
}
