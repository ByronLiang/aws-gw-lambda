package model

type ResizeConfig struct {
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Bucket   string `json:"bucket"`
	Region   string `json:"region"`
	FileName string `json:"fileName"`
}
