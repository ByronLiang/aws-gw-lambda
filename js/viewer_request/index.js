'use strict';

const querystring = require('querystring');

// 查看器请求/源请求触发函数: 实现请求动态缩放的尺寸不属于规定的尺寸，则采用就近原则，使用常用尺寸中最接近的尺寸

// 指定常规的尺寸值
const allowedDimension = [ {w:100,h:100}, {w:200,h:200}, {w:300,h:300}, {w:400,h:400} ];
// 尺寸容错比值
const variance = 20;
// 是否校验请求尺寸
const vertifyDimension = false;

exports.handler = (event, context, callback) => {
    const request = event.Records[0].cf.request;
    console.log(request.uri)
    let key = decodeURIComponent(request.uri).substring(1);
    // 解析参数
    let res = parseQuery(key);
    let originFileKey = res.originFileKey;
    let width = res.width;
    let height = res.height;
    // 文件类型
    let format = res.format;
    // 不符合正则规则, 不处理
    if (! res.canResizeImage) {
        callback(null, request);
        return;
    }
    if (vertifyDimension) {
        let variancePercent = (variance/100);
        let matchFound = false;
        for (let dimension of allowedDimension) {
            let minWidth = dimension.w - (dimension.w * variancePercent);
            let maxWidth = dimension.w + (dimension.w * variancePercent);
            if(width >= minWidth && width <= maxWidth){
                width = dimension.w;
                height = dimension.h;
                matchFound = true;
                break;
            }
        }
        if (!matchFound) {
            // 无法匹配, 从常规尺寸边缘进行匹配
            let maxDimension = allowedDimension[allowedDimension.length - 1]
            if (maxDimension.w > width) {
                // 取常规尺寸的最小值
                width = allowedDimension[0].w;
                height = allowedDimension[0].h;
            } else {
                // 取最大常规尺寸值
                width = maxDimension.w;
                height = maxDimension.h;
            }
        }
    }
    // let queryStr = querystring.stringify({size: width + "x" + height});
    // queryStr = "?" + queryStr + "." + format;
    // let encodeQuery = encodeURIComponent(queryStr)
    // request.uri = "/" + originFileKey + encodeQuery;
    let filePath = originFileKey + "_" + width + "x" + height + "." + format;
    request.uri = "/" + filePath;
    callback(null, request);

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
        return {canResizeImage: true, originFileKey: originFileKey, width: width, height: height, format: format};
    }
};