package util

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"image/gif"
	"log"
	"os"

	"github.com/ByronLiang/aws-gw-lambda/model"
	"github.com/disintegration/imaging"
)

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
