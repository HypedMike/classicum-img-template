package main

import (
	"encoding/json"
	"fmt"
	"img-template/lib/img"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type optionsStruct struct {
	Label           string   `json:"label"`
	LabelColor      string   `json:"label_color"`
	BackgroundImage string   `json:"background_image"`
	Logos           []string `json:"logos"`
	FontPath        string   `json:"font_path"`
	FontSize        int      `json:"font_size"`
	SavePath        string   `json:"save_path"`
}

func main() {
	optionsPath := os.Args[1]

	// check if the file exists
	if _, err := os.Stat(optionsPath); os.IsNotExist(err) {
		fmt.Println("File does not exist")
		os.Exit(1)
	}

	// read file
	file, err := os.ReadFile(optionsPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// parse the file into optionsStruct as a json
	var options optionsStruct
	err = json.Unmarshal([]byte(file), &options)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// create the image
	createImage(options)

}

func createImage(options optionsStruct) {
	image := img.NewImg(1000, 1000, 1000)
	err := image.AddBackgroundImageFromPath(options.BackgroundImage)
	if err != nil {
		fmt.Println(err)
	}

	// convert hex to rgba
	rgba, err := hexToRGBA(options.LabelColor)
	if err != nil {
		fmt.Println(err)
	}

	color := img.ColorRGBA{
		A: uint8(rgba[3]),
		R: uint8(rgba[0]),
		G: uint8(rgba[1]),
		B: uint8(rgba[2]),
	}

	err = image.AddTextCentral(options.Label, color, options.FontSize)
	if err != nil {
		fmt.Println(err)
	}
	err = image.SaveImage(fmt.Sprintf("%s/%s.png", options.SavePath, uuid.New().String()))
	if err != nil {
		fmt.Println(err)
	}
}

func hexToRGBA(hex string) ([]int, error) {
	var rgba []int

	// Remove the "#" prefix
	hex = strings.TrimPrefix(hex, "#")

	// Convert the hex string to an integer
	hexInt, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return []int{}, fmt.Errorf("invalid hex color: %s", hex)
	}

	// Extract the red, green, and blue components
	rgba = append(rgba, int(hexInt>>16))
	rgba = append(rgba, int((hexInt>>8)&0xff))
	rgba = append(rgba, int(hexInt&0xff))
	rgba = append(rgba, 255)

	// Return the RGBA color
	return rgba, nil
}
