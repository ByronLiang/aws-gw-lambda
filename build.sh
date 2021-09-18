rm main.zip
GOARCH=amd64 GOOS=linux go build -o main cmd/imageResize/main.go
# Window打包
build-lambda-zip -output main.zip main
rm main