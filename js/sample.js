'use strict';

const sharp = require('sharp');

let crop = async () => {
    await sharp('sample.jpg')
    .resize({
        width: 400,
        height: 300,
        fit: 'inside',
    })
    .sharpen()
    .toFile('output.jpg')
    .then(info => { 
        console.log(info);
    })
    .catch(err => {
        console.log(err);
    });
};

let compress = async () => {
    await sharp('sample.jpg').jpeg({
        quality: 40,
    })
    .toFile('compress_output.jpg')
    .then(info => { 
        console.log(info);
    })
    .catch(err => {
        console.log(err);
    });
}

// 裁剪demo
crop();

// 压缩demo
compress();

