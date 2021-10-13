'use strict';

const aws = require('aws-sdk');
const sharp = require('sharp');

const region = "us-east-1";
const bucket = "byronbook";
// 支持裁剪的文件类型
const supportImageTypes = ['jpg', 'jpeg', 'png'];
// 指定可裁剪的尺寸值
// const allowedDimension = [ {w:100,h:100}, {w:200,h:200}, {w:300,h:300}, {w:400,h:400} ];
// 测试不限制
const allowedDimension = [];

const s3 = new aws.S3({
  region: region,
  signatureVersion: 'v4'
});

// const querystring = require('querystring');
exports.handler = async (event, context, callback) => {
    let canResizeImage = false;
    //Get contents of response
    const response = event.Records[0].cf.response;
    const request = event.Records[0].cf.request;
    if (response.status == 404 || response.status == 403) {
        let key = decodeURIComponent(request.uri).substring(1);
        // console.log("key", key)
        // console.log("response \n", JSON.stringify(response))
        // 参数示例: "images/ims-web/08bf254d-f8b4-4711-88cf-37390f00dd27.jpg_200x200.jpg"
        // 解析参数
        let mathcGroup = key.match(/(.*)_(\d+)x(\d+)\.(.*)/);
        // 不符合正则规则, 不处理
        if (mathcGroup === null) {
            callback(null, response);
            return;
        }
        // 解析异常, 不处理
        if (mathcGroup.length !== 5) {
            callback(null, response);
            return;
        }
        let newFileKey = mathcGroup[0];
        // 原文件key
        let originFileKey = mathcGroup[1];
        let width = parseInt(mathcGroup[2], 10);
        let height = parseInt(mathcGroup[3], 10);
        // 文件类型
        let format = mathcGroup[4].toLowerCase();
        // 校验文件类型是否符号自定义裁剪
        let isSupportImageFormat = supportImageTypes.some(type => {
            return type == format;
        });
        if (isSupportImageFormat) {
            // 没配置指定尺寸, 默认都进行裁剪处理
            if (allowedDimension.length === 0) {
                canResizeImage = true;
            } else {
                // 校验尺寸
                for (let dimension of allowedDimension) {
                    if (dimension.w === width && dimension.h === height) {
                        canResizeImage = true;
                        break
                    }
                }
            }
        }
        // 不符合裁剪校验, 返回响应
        if (!canResizeImage) {
            callback(null, response);
            return;
        }
        try {
            // get the source image file
            const s3Object = await s3
                .getObject({
                    Bucket: bucket,
                    Key: originFileKey
                })
                .promise();
            // 获取图片内容异常
            if (s3Object.ContentLength == 0) {
                callback(null, response);
                return;
            }
            let imageObj, metaData, buffer
            imageObj = await sharp(s3Object.Body).rotate();
            metaData = await imageObj.metadata();
            // 解析裁剪参数, 进行裁剪处理
            if (metaData.width > width || metaData.height > height) {
                console.log("start resize image");
                imageObj.resize(width, height);
            }
            buffer = await imageObj.toBuffer();
            console.log("start to upload")
            // 异步上传到s3
            s3.putObject({
                Body: buffer,
                Bucket: bucket,
                ContentType: 'image/' + format,
                Key: newFileKey,
            }).promise().catch((err) => { 
                console.log("Exception while writing resized image to bucket", newFileKey);
                console.error(err);
            });
            console.log("end upload to s3");
            response.status = 200;
            response.body = buffer.toString('base64');
            response.bodyEncoding = 'base64';
            response.headers['content-type'] = [{ key: 'Content-Type', value: 'image/' + format }];
            callback(null, response);
            return
        } catch (err) {
            console.log("catch error");
            console.error(err);
            callback(null, response);
            return;
        }
    }
    //Return modified response
    callback(null, response);
};
