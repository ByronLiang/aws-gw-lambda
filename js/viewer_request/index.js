'use strict';

// 查看器请求触发函数: 实现请求动态缩放的尺寸不属于规定的尺寸，则采用就近原则，使用常用尺寸中最接近的尺寸

// 指定常规的尺寸值
const allowedDimension = [ {w:100,h:100}, {w:200,h:200}, {w:300,h:300}, {w:400,h:400} ];
// 尺寸容错比值
const variance = 20;

exports.handler = (event, context, callback) => {
    const request = event.Records[0].cf.request;
    console.log(request.uri)
    let key = decodeURIComponent(request.uri).substring(1);
    // 参数示例: "images/ims-web/08bf254d-f8b4-4711-88cf-37390f00dd27.jpg_200x200.jpg"
    // 解析参数
    let mathcGroup = key.match(/(.*)_(\d+)x(\d+)\.(.*)/);
    // 不符合正则规则, 不处理
    if (mathcGroup === null) {
        callback(null, request);
        return;
    }
    // 解析异常, 不处理
    if (mathcGroup.length !== 5) {
        callback(null, request);
        return;
    }
    let newFileKey = mathcGroup[0];
    // 原文件key
    let originFileKey = mathcGroup[1];
    let width = parseInt(mathcGroup[2], 10);
    let height = parseInt(mathcGroup[3], 10);
    // 文件类型
    let format = mathcGroup[4].toLowerCase();

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
    request.uri = "/" + originFileKey + "_" + width + "x" + height + "." + format;
    callback(null, request);
};