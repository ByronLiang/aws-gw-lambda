package config

import "os"

var ResizeImageLambdaConfig resizeImageLambdaConfig

type resizeImageLambdaConfig struct {
	Region     string
	Bucket     string
	BucketUrl  string
	PathPrefix string
}

func init() {
	ResizeImageLambdaConfig = resizeImageLambdaConfig{
		Region:     os.Getenv("region"),
		Bucket:     os.Getenv("bucket"),
		BucketUrl:  os.Getenv("bucket_url"),
		PathPrefix: os.Getenv("path_prefix"),
	}
}
