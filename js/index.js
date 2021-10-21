'use strict';

const aws = require('aws-sdk');
const sharp = require('sharp');
const querystring = require('querystring');

const region = "us-east-1";
const bucket = "byronbook";
// 支持裁剪的文件类型
const supportImageTypes = ['jpg', 'jpeg', 'png'];
// 指定可裁剪的尺寸值
// const allowedDimension = [ {w:100,h:100}, {w:200,h:200}, {w:300,h:300}, {w:400,h:400} ];

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
        // 解析参数
        let res = parseQuery(key);
        let newFileKey = key;
        let originFileKey = res.originFileKey;
        let width = res.width;
        let height = res.height;
        let format = res.format;
        canResizeImage = res.canResizeImage;
        // 不符合裁剪校验, 返回响应
        if (!canResizeImage) {
            callback(null, response);
            return;
        }
        let filePath = genFileName(originFileKey, format, width, height);
        // 查看bucket 是否存在此文件
        let filePathRes = await getObjectFromS3(filePath);
        if (filePathRes.objExist) {
            // 源文件存在
            response.status = 200;
            response.body = filePathRes.data.Body.toString('base64');
            response.bodyEncoding = 'base64';
            response.headers['content-type'] = [{ key: 'Content-Type', value: 'image/' + format }];
            callback(null, response);
            return
        }
        // 使用原文件进行裁剪图片
        let originFileRes = await getObjectFromS3(originFileKey);
        // 原生尺寸图片文件不存在
        if (!originFileRes.objExist) {
            callback(null, response);
            return;
        }
        // 开始裁剪图片
        let imageObj, metaData, buffer
        imageObj = await sharp(originFileRes.data.Body).rotate();
        metaData = await imageObj.metadata();
        // 解析裁剪参数, 进行裁剪处理
        // fit: inside 保持纵横比裁剪
        if (metaData.width > width || metaData.height > height) {
            imageObj.resize(width, height, { fit: 'inside' });
        }
        buffer = await imageObj.toBuffer();
        // 自定义生成s3的文件名
        newFileKey = genFileName(originFileKey, format, width, height);
        console.log("start to upload", newFileKey);
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
        response.status = 200;
        response.body = buffer.toString('base64');
        response.bodyEncoding = 'base64';
        response.headers['content-type'] = [{ key: 'Content-Type', value: 'image/' + format }];
        callback(null, response);
        return
    }
    //Return modified response
    callback(null, response);

    function parsePath(key) {
        // 参数示例: "images/ims-web/08bf254d-f8b4-4711-88cf-37390f00dd27.jpg_200x200.jpg"
        // 解析参数
        let mathcGroup = key.match(/(.*)_(\d+)x(\d+)\.(.*)/);
        // 不符合正则规则, 不处理
        if (mathcGroup === null) {
            return {canResizeImage: false, originFileKey: "", width: 0, height: 0, format: ""};
        }
        // 解析异常, 不处理
        if (mathcGroup.length !== 5) {
            return {canResizeImage: false, originFileKey: "", width: 0, height: 0, format: ""};
        }
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
            return {canResizeImage: true, originFileKey: originFileKey, width: width, height: height, format: format};
        } else {
            return {canResizeImage: false, originFileKey: "", width: 0, height: 0, format: ""};
        }
    }

    function parseQuery(key) {
        // 参数示例: "images/ims-web/08bf254d-f8b4-4711-88cf-37390f00dd27.jpg?size=200x200"
        // 解析参数
        let mathcGroup = key.match(/(.*\.(.*))\?(.*)/);
        // 不符合正则规则, 不处理
        if (mathcGroup === null) {
            return {canResizeImage: false, originFileKey: "", width: 0, height: 0, format: ""};
        }
        // 解析异常, 不处理
        if (mathcGroup.length !== 4) {
            return {canResizeImage: false, originFileKey: "", width: 0, height: 0, format: ""};
        }
        // 原文件key
        let originFileKey = mathcGroup[1];
        // 文件类型
        let format = mathcGroup[2].toLowerCase();
        // 参数请求
        let query = mathcGroup[3];
        let params = querystring.parse(query);
        if (!params.size) {
            return {canResizeImage: false, originFileKey: "", width: 0, height: 0, format: ""};
        }
        let sizeGroup = params.size.split("x");
        if (sizeGroup.length !== 2) {
            return {canResizeImage: false, originFileKey: "", width: 0, height: 0, format: ""};
        }
        let width = parseInt(sizeGroup[0], 10);
        let height = parseInt(sizeGroup[1], 10);
        // 校验文件类型是否符号自定义裁剪
        let isSupportImageFormat = supportImageTypes.some(type => {
            return type == format;
        });
        if (isSupportImageFormat) {
            return {canResizeImage: true, originFileKey: originFileKey, width: width, height: height, format: format};
        }
        return {canResizeImage: false, originFileKey: "", width: 0, height: 0, format: ""};
    }

    function genFileName(originFileKey, format, width, height) {
        return originFileKey + "_" + width + "x" + height + "." + format
    }

    async function getObjectFromS3(path) {
        var params = {
            Bucket: bucket,
            Key: path
        };
        try {
            const s3Object = await s3.getObject(params).promise();
            return {objExist: true, data: s3Object};
        } catch (err) {
            return {objExist: false, data: ""};
        }
    }
};
