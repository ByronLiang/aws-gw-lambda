'use strict';

const aws = require('aws-sdk');
const sharp = require('sharp');

const region = "us-east-1";
const bucket = "byronbook";
const supportImageTypes = ['jpg', 'jpeg', 'png'];
const allowedDimension = [ {w:100,h:100}, {w:200,h:200}, {w:300,h:300}, {w:400,h:400} ];
const resizeType = "cover";

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
        console.log("key", key)
        console.log("response \n", JSON.stringify(response))
        // 解析参数
        let mathcGroup = key.match(/(.*)_(\d+)x(\d+)(.*)\.(.*)/);
        if (mathcGroup === null) {
            callback(null, response);
            return;
        }
        if (mathcGroup.length !== 6) {
            callback(null, response);
            return;
        }
        let newFileKey = mathcGroup[0];
        // 原文件key
        let originFileKey = mathcGroup[1];
        let width = parseInt(mathcGroup[2], 10);
        let height = parseInt(mathcGroup[3], 10);
        let filename = mathcGroup[4];
        // 文件类型
        let format = mathcGroup[5].toLowerCase();
        // 校验文件类型是否符号自定义裁剪
        let isSupportImageFormat = supportImageTypes.some(type => {
            return type == format;
        });
        if (isSupportImageFormat) {
            // 校验尺寸
            for (let dimension of allowedDimension) {
                if (dimension.w === width && dimension.h === height) {
                    canResizeImage = true;
                    break
                }
            }
        }
        try {
            // get the source image file
            const s3Object = await s3
                .getObject({
                    Bucket: bucket,
                    Key: originFileKey
                })
                .promise();
            if (s3Object.ContentLength == 0) {
                callback(null, response);
                return;
            }
            let imageObj, metaData, buffer
            // imageObj = await sharp(s3Object.Body).rotate();
            if (canResizeImage) {
                imageObj = await sharp(s3Object.Body).rotate();
                metaData = await imageObj.metadata();
                // 解析裁剪参数, 进行裁剪处理
                if (metaData.width > width || metaData.height > height) {
                    console.log("start resize image");
                    imageObj.resize(width, height);
                }
                buffer = await imageObj.toBuffer();
                let byteLength = Buffer.byteLength(buffer, "base64");
                console.log("buffer length: ", byteLength)
                console.log("start to upload")
                // upload to s3
                const s3PutObjectRes = await s3.putObject({
                    Body: buffer,
                    Bucket: bucket,
                    ContentType: 'image/' + format,
                    Key: newFileKey,
                }).promise()
                console.log("end upload to s3")
            } else {
                console.log("no resize image")
                buffer = s3Object.Body
                // buffer = await imageObj.toBuffer();
            }
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
    
    function responseUpdate(
        status,
        statusDescription,
        body,
        contentHeader,
        bodyEncoding = undefined
    ) {
        response.status = status;
        response.statusDescription = statusDescription;
        response.body = body;
        response.headers["content-type"] = contentHeader;
        if (bodyEncoding) {
          response.bodyEncoding = bodyEncoding;
        }
    }
};
