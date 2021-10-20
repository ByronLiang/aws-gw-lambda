# CloudFront Lambda Edge NodeJs 图片处理

## 编译

使用Node12进行编译, `npm install`

在项目根目录, 打包zip文件。Linux 环境下, `zip -r index.zip .`

## 图片处理库

1. sharp 库
2. npm install 比较复杂，涉及环境安装libvips 二进制包
3. 常见错误: `prebuild-install WARN install No prebuilt binaries found` Node14 需要安装 sharp 版本0.25.2 以上

## 配置

### 函数程序配置

1. `const region = ` 配置region

2. `const bucket = ` 配置bucket

### aws lambda 函数配置

1. 设置lambda 函数, 触发器选择 cloudfront，相关权限配置: edge 执行权限，上传s3权限
2. 每当更新函数与变更函数配置，需要更新版本，并对指定的cloudfront部署, 并需要全部节点都完成部署，才能生效
3. 选择触发行为: 查看者请求，查看者响应，源请求，源响应; 一般是源响应

## 调试与日志

1. 需要到指定的region 查看日志组
2. 在指定的cloudfront里查看 lambda edge 执行次数与所在区域(region)，从而定位日志组
3. console.log 打点日志

## 参数解析

1. 若请求参数是以query, query里的key 与 value 必须先进行编码`(encodeURIComponent)`
2. query的 parameter 顺序会影响cloudfront命中
