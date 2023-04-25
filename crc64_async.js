/**
 * JS 异步并发读取文件，分片计算 crc64，最后合起来
 */
const { crc64File } = require("./crc64_ecma182");
const { crc64_concat } = require("./crc64");
const { eachLimit } = require("async");
const fs = require("fs");

function file_crc64(filePath, callback) {
    var size = fs.statSync(filePath).size;
    var maxCount = 4;
    var chunkSize = Math.ceil(size / maxCount);
    var chunkCount = Math.ceil(size / chunkSize);
    var chunkList = [];
    var start = 0, end;
    for (var i = 0; i < chunkCount; i++) {
        end = Math.min(size - 1, start + chunkSize - 1);
        chunkList.push({start, end});
        start += chunkSize;
    }
    var hashList = [];
    eachLimit(chunkList, chunkCount, (item, next) => {
        var { start, end } = item;
        crc64File({filePath, start, end}, (err, hash) => {
            var size = end - start + 1;
            hashList.push({hash, size});
            console.log(filePath, start, end, hash);
            next();
        });
    }, (err) => {
        if (err) return callback(err);
        var finalHash = crc64_concat(hashList);
        callback(null, finalHash);
    });
}

module.exports = file_crc64;

if (!module.parent) {
    var time0 = Date.now();
    var filePath = '1g.zip';
    var hash = file_crc64(filePath, (err, hash) => {
        var time1 = Date.now();
        console.log(filePath, hash, (time1 - time0) / 1000 + 's');
    });
    // 1g.zip 3534425600523290380n 1.991s
    // 10g.zip 4594563571287695306n 18.087s
}