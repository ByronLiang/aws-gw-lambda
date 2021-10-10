# CloudFront Lambda Edge NodeJs 图片处理

## 图片处理库
1. sharp 库
2. npm install 比较复杂，涉及环境安装libvips 二进制包

## aws 配置
1. 设置lambda 函数, 触发器选择 cloudfront，相关权限配置: edge 执行权限，上传s3权限
2. 每当更新函数与变更函数配置，需要更新版本，并对指定的cloudfront 部署
3. 正确选择响应，一般是源响应

## 调试与日志
1. 需要到指定的region 查看日志组
2. 在指定的cloudfront里查看 lambda edge 执行次数与所在区域(region)，从而定位日志组
3. console.log 打点日志
