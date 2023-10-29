package img

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/freetype/truetype"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
	fon "golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Img struct {
	// sizes
	Width  int
	Height int
	Dpi    int
	// image
	Base draw.Image
}

func NewImg(width, height int, dpi int) *Img {
	//create image’s background
	bgImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// set the background color
	bgColor := image.Uniform{C: image.Transparent.C}
	draw.Draw(bgImg, bgImg.Bounds(), &bgColor, image.Point{}, draw.Src)

	return &Img{
		Width:  width,
		Height: height,
		Base:   bgImg,
		Dpi:    dpi,
	}
}

func (i *Img) AddImage(img image.Image, x, y int) {
	draw.Draw(i.Base, img.Bounds().Add(image.Pt(x, y)), img, image.Point{}, draw.Over)
}

/*
add text to image
*/
func (img *Img) addText(x, y int, label string, _color ColorRGBA, fontPath string) error {
	col := color.RGBA{_color.R, _color.G, _color.B, _color.A}
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}

	// Load the font file
	fontBytes, err := os.ReadFile("/Users/michelesaladino/Documents/code/classicum-img-template/lib/fonts/short-baby-font/ShortBaby-Mg2w.ttf")
	if err != nil {
		return err
	}

	// Parse the font file
	font, err := truetype.Parse(fontBytes)
	if err != nil {
		return err
	}

	f := truetype.NewFace(font, &truetype.Options{
		Size: 100,
	})

	d := &fon.Drawer{
		Dst:  img.Base,
		Src:  image.NewUniform(col),
		Face: f,
		Dot:  point,
	}
	d.DrawString(label)
	return nil
}

func (img *Img) AddTextCentral(label string, _color ColorRGBA, fontSize int, fontPath string) error {

	// if text length exceeds image width, divide text
	if len(label)*fontSize > img.Width {
		labels := img.getStrings(label, fontSize)
		for i, l := range labels {
			newY := img.Height/2 - (fontSize * len(labels) / 2) + (i * fontSize) + 50
			newX := (img.Width / 2) - (len(l) * fontSize / 4) + 50
			err := img.addText(newX, newY, l, _color, fontPath)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// calculate central point in image given the font size
	x := (img.Width / 2) - (len(label) * fontSize / 4)
	y := (img.Height / 2) - (fontSize / 2)

	img.addText(x, y, label, _color, fontPath)
	return nil
}

/*
add background image
*/
func (i *Img) AddBackgroundImageFromPath(path string) error {
	// check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	// get image size
	size, err := getImgSize(path)
	if err != nil {
		return err
	}

	// reformat image size
	width := i.Width
	ratio := float64(size[0]) / float64(size[1])
	height := float64(width) / ratio

	// scale picture
	path, err = scaleImage(path, uint(width), uint(height))
	if err != nil {
		return err
	}

	// crop picture
	path, err = cropImage(path, 0, 0, i.Width, i.Height)
	if err != nil {
		return err
	}

	// Open the JPEG file
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the JPEG image
	img, err := jpeg.Decode(file)
	if err != nil {
		return err
	}

	// Draw the image onto the base image covering the whole base
	draw.Draw(i.Base, image.Rect(0, 0, 1000, 1000).Bounds(), img, image.Point{}, draw.Over)

	return nil
}

func (i *Img) SaveImage(filename string) error {

	// get extension
	ext := strings.ToLower(filepath.Ext(filename))

	// init error
	var err error

	// create file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	// save image
	switch ext {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(file, i.Base, &jpeg.Options{Quality: 100})
	case ".png":
		err = png.Encode(file, i.Base)
	default:
		err = errors.New("unsupported file format")
	}
	defer file.Close()
	return err
}

/*
add logos to bottom of image
*/
func (i *Img) AddLogos(filenames []string) error {
	// get size per logo
	size := i.Width / len(filenames)

	// scale and crop logos
	var croppedLogos []string
	for _, filename := range filenames {
		croppedLogo, err := scaleAndCrop(filename, size, size)
		if err != nil {
			return err
		}
		croppedLogos = append(croppedLogos, croppedLogo)
	}

	// add logos to image
	for index, croppedLogo := range croppedLogos {
		i.AddImage(imageFromPath(croppedLogo), index*size, i.Height-size)
	}

	return nil
}

/*
gets array of strings based on image width and font size
*/
func (i *Img) getStrings(label string, fontSize int) []string {
	// init array of strings
	var stringArray []string

	// image width
	width := i.Width

	// calculate max chars per line
	maxCharsPerLine := width / fontSize

	// words in label
	words := strings.Split(label, " ")

	// clean words
	for i, w := range words {
		words[i] = removeSpaces(w)
	}

	var currentString []string
	for i := 0; i < len(strings.Split(label, " ")); i++ {
		if len(strings.Join(currentString, " ")) < maxCharsPerLine {
			currentString = append(currentString, words[i])
		} else {
			stringArray = append(stringArray, strings.Join(currentString, " "))
			currentString = []string{words[i]}
		}

		if i == len(words)-1 {
			stringArray = append(stringArray, strings.Join(currentString, " "))
		}
	}

	return stringArray
}

/*
removes spaces before and after string
*/
func removeSpaces(label string) string {
	return strings.TrimSpace(label)
}

func scaleImage(path string, width, height uint) (string, error) {
	// Open the image file
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	// Resize the image
	resized := resize.Resize(width, height, img, resize.Lanczos3)

	tempPath := fmt.Sprintf("%s/%s.%s", os.TempDir(), uuid.New().String(), "jpg")
	fmt.Println(tempPath)

	// Save the resized image to a file
	out, err := os.Create(tempPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write the resized image to the file
	jpeg.Encode(out, resized, nil)

	return tempPath, nil
}

func getImgSize(path string) ([]int, error) {
	// Open the image file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode the image config
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return nil, err
	}

	// Return the image size as a slice of integers
	return []int{config.Width, config.Height}, nil
}

func cropImage(path string, x, y, width, height int) (string, error) {
	// Open the image file
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	tempPath := fmt.Sprintf("%s/%s.%s", os.TempDir(), uuid.New().String(), "jpg")

	// Crop the image
	cropped := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(x, y, x+width, y+height))

	// Save the cropped image to a file
	out, err := os.Create(tempPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write the cropped image to the file
	jpeg.Encode(out, cropped, nil)

	return tempPath, nil
}

/*
scale and crop images automatically
*/
func scaleAndCrop(path string, w, h int) (string, error) {
	newpath, err := scaleImage(path, uint(w), uint(h))
	if err != nil {
		return "", err
	}
	return cropImage(newpath, 0, 0, w, h)
}

func imageFromPath(path string) image.Image {
	// Open the image file
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Println(err)
	}

	return img
}
