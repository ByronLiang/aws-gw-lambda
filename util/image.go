package util

import (
	"bytes"
	"log"

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
