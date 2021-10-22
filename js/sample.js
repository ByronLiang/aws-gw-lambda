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

let getClosestDimension = (width) => {
    let table = [ {w:100,h:100}, {w:275,h: 200}, {w:300,h:300}, {w:400,h:400} ];;
    table.sort((a, b) => a.w - b.w);
    console.log(table);
    let l = 0;
    let r = table.length - 1;
    let m = 0;
    let match = false;
    while (l <= r) {
        m = l + parseInt((r - l) >> 1);
        if (table[m].w == width) {
            match = true;
            l = m;
            break
        }
        if (table[m].w > width) {
            r = m - 1;
        } else {
            l = m + 1;
        }
    }
    // 取最小尺寸值
    if (l === 0) {
        return table[l];
    }
    // 取最大尺寸值
    if (l === table.length) {
        return table[table.length - 1];
    }
    if (!match) {
        return table[l-1];
    }
    return table[l];
}

// 裁剪demo
// crop();

// 压缩demo
// compress();

let dimension = getClosestDimension(190);
console.log(dimension.w, dimension.h);

