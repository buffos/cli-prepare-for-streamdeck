# CLI Prepare for StreamDeck

## Overview

This CLI tool is designed to help prepare media files for use with StreamDeck, a customizable keyboard that allows users to trigger actions and display custom icons. The application provides a user-friendly interface to process and organize media files (images and videos) for StreamDeck integration.

## Features

- Interactive CLI interface built with Bubble Tea
- Media folder preparation for StreamDeck
- Image and video thumbnail generation
- Customizable border colors for thumbnails
- OSC (Open Sound Control) path configuration
- Support for multiple media types

## Main Functionality

### Media Preparation

The tool offers the following key functions:

1. **Prepare Media Folder**:

   - Processes images and videos in a specified directory
   - Generates thumbnails with configurable border colors
   - Creates pressed state images for interactive buttons
   - Generates a JSON configuration file for StreamDeck integration

2. **Echo Command**:
   - A placeholder for future development
   - Can be used to test the CLI tool's functionality

### Configuration

The application uses a configuration system that allows you to:

- Specify media types (images, videos)
- Set OSC root paths
- Define border colors and widths for thumbnails

## Requirements

- Go 1.16+
- FFmpeg (for video thumbnail extraction)

## Installation

```bash
go get github.com/buffos/cli-prepare-for-streamdeck
go build
```

## Usage

Run the application and select from the main menu:

- Prepare Media Folder
- Echo Command
- Quit

## Dependencies

- Bubble Tea (github.com/charmbracelet/bubbletea)
- Lipgloss (github.com/charmbracelet/lipgloss)
- Imaging (github.com/disintegration/imaging)

## License

MIT License

## Contributing

Contributions are welcome! Please submit pull requests or open issues on the GitHub repository.
