'use strict';

const sharp = require('sharp');

let test = async () => {
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

test();

