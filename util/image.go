package util

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strings"

	exifremove "github.com/scottleedavis/go-exif-remove"

	"github.com/ByronLiang/aws-gw-lambda/model"
	"github.com/disintegration/imaging"
)

type Format int

// Image file formats.
const (
	JPEG Format = iota
	PNG
	GIF
	TIFF
	BMP
)

var formatExts = map[string]Format{
	"jpg":  JPEG,
	"jpeg": JPEG,
	"png":  PNG,
	"gif":  GIF,
	"tif":  TIFF,
	"tiff": TIFF,
	"bmp":  BMP,
}

func GetResizedImage(resizeConfig *model.ResizeConfig) ([]byte, error) {
	imageObj, err := imaging.Decode(bytes.NewReader(resizeConfig.ImageByte))
	if err != nil {
		log.Printf("imaging decode source image error: %s", err.Error())
		return nil, err
	}
	newImageObj := imaging.Resize(imageObj, resizeConfig.Width, resizeConfig.Height, imaging.Lanczos)
	buf := new(bytes.Buffer)
	err = imaging.Encode(buf, newImageObj, resizeConfig.FileFormat)
	if err != nil {
		log.Printf("imaging encode new image byte error: %s", err.Error())
		return nil, err
	}
	return buf.Bytes(), nil
}

func GetResizeGif(resizeConfig *model.ResizeConfig) ([]byte, error) {
	if resizeConfig.FileFormat != imaging.GIF {
		log.Println("file format is not gif")
		return nil, errors.New("file format error")
	}
	gifList, err := gif.DecodeAll(bytes.NewReader(resizeConfig.ImageByte))
	if err != nil {
		log.Println("gif decode all error")
		return nil, err
	}
	for i, gifData := range gifList.Image {
		newImg := imaging.Resize(gifData, resizeConfig.Width, resizeConfig.Height, imaging.Lanczos)
		gifList.Image[i] = image.NewPaletted(
			image.Rect(0, 0, resizeConfig.Width, resizeConfig.Height),
			gifList.Image[i].Palette)
		draw.Draw(gifList.Image[i],
			image.Rect(0, 0, resizeConfig.Width, resizeConfig.Height),
			newImg,
			image.Pt(0, 0), draw.Src)
	}
	buf := new(bytes.Buffer)
	err = gif.EncodeAll(buf, gifList)
	if err != nil {
		log.Println("gif encode all error")
		return nil, err
	}
	return buf.Bytes(), nil
}

// 翻转 Gif 图
func ReserveGif(gifByte []byte, filename string) error {
	gifList, err := gif.DecodeAll(bytes.NewReader(gifByte))
	if err != nil {
		log.Println("gif decode all error")
		return err
	}
	i := 0
	j := len(gifList.Image) - 1
	for i < j {
		gifList.Image[i], gifList.Image[j] = gifList.Image[j], gifList.Image[i]
		i++
		j--
	}
	outputFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	err = gif.EncodeAll(outputFile, gifList)
	if err != nil {
		return err
	}
	return nil
}

func CompressImageResource(data []byte, quality int) []byte {
	imgSrc, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return data
	}
	newImg := image.NewRGBA(imgSrc.Bounds())
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(newImg, newImg.Bounds(), imgSrc, imgSrc.Bounds().Min, draw.Over)
	buf := bytes.Buffer{}
	err = jpeg.Encode(&buf, newImg, &jpeg.Options{Quality: quality})
	if err != nil {
		return data
	}
	if buf.Len() > len(data) {
		return data
	}
	return buf.Bytes()
}

func GetImageExif(data []byte) ([]byte, error) {
	return exifremove.Remove(data)
	//png := pngstructure.NewPngMediaParser()
	//jmp := jpegstructure.NewJpegMediaParser()
	// return nil, nil
}

func FormatFromFilename(filename string) Format {
	ext := filepath.Ext(filename)
	if f, ok := formatExts[strings.ToLower(strings.TrimPrefix(ext, "."))]; ok {
		return f
	}
	return -1
}

// 压缩图片并返回图片字节数据
func GetImageCompressByte(img image.Image, format Format, quality int) ([]byte, error) {
	var err error
	buf := new(bytes.Buffer)
	switch format {
	case JPEG:
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case PNG:
		newImg := image.NewRGBA(img.Bounds())
		draw.Draw(newImg, newImg.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
		draw.Draw(newImg, newImg.Bounds(), img, img.Bounds().Min, draw.Over)
		err = jpeg.Encode(buf, newImg, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	return nil, nil
}
