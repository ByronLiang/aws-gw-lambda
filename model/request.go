package model

import "github.com/disintegration/imaging"

type ResizeConfig struct {
	Width      int
	Height     int
	FileName   string
	FileFormat imaging.Format
	ImageByte  []byte
}
