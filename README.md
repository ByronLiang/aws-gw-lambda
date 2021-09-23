# AWS Gateway Lambda 调用

利用静态资源重定向规则，对自定义的路径触发重定向API Gateway，状态码为307，是GET请求，从而执行触发器里的程序逻辑

仓库代码主要呈现触发器执行的工程文件

1. 主要涉及获取重定向的请求参数

2. 响应：可以是状态码为301的重定向，呈现图片资源。也可以无状态响应

## 基本编译

将工程编译二进制文件，并进行zip文件打包，上传到AWS的触发器代码源处

### 参考

 [aws api gateway 官方](https://aws.amazon.com/cn/blogs/compute/resize-images-on-the-fly-with-amazon-s3-aws-lambda-and-amazon-api-gateway/)

 [aws-s3-lambda-api-gateway-at-golang](https://medium.com/@ducmeit/build-a-resize-images-tool-with-aws-s3-lambda-api-gateway-at-golang-7569c72c3e8a)

 落地:

 [gw-resize-image-tool](https://github.com/ducmeit1/golang-resize-image-tool)

 开发难点:

 [gif resize issue](https://github.com/disintegration/imaging/issues/23)
 [Cloud Front cahce issue](https://github.com/sagidM/s3-resizer/issues/5)
