/**
 * JS Worker 并发读取文件，分片计算 crc64，最后合起来
 */
const { crc64File } = require("./crc64_ecma182");
const { crc64_concat } = require("./crc64");
const worker_threads = require("worker_threads");
const { eachLimit } = require("async");
const pathLib = require("path");
const fs = require("fs");
const os = require("os");

var MAX_THREAD = Math.max(3, Math.min(os.cpus().length));
function file_crc64(filePath, callback) {
    var size = fs.statSync(filePath).size;
    var chunkSize = Math.ceil(size / MAX_THREAD);
    var chunkCount = Math.ceil(size / chunkSize);
    var chunkList = [];
    var start = 0, end;
    for (var i = 0; i < chunkCount; i++) {
        end = Math.min(size - 1, start + chunkSize - 1);
        chunkList.push({index: i, start, end});
        start += chunkSize;
    }
    var hashList = Array(chunkCount);
    eachLimit(chunkList, chunkCount, (item, next) => {
        var { index, start, end } = item;
        const { Worker } = worker_threads;
        let seprateThread = new Worker(__filename);
        seprateThread.on("message", (hash) => {
            var size = end - start + 1;
            hashList[index] = {hash, size};
            seprateThread.terminate();
            seprateThread = null;
            next();
        });
        var opt = {filePath: pathLib.resolve(filePath), start, end};
        seprateThread.postMessage(opt);
    }, (err) => {
        if (err) return callback(err);
        var finalHash = crc64_concat(hashList);
        callback(null, finalHash);
    });
}

module.exports = file_crc64;

// 子线程
if (worker_threads && worker_threads.parentPort) {
    const { parentPort } = worker_threads;
    parentPort.on("message", (opt) => {
        const {filePath, start, end} = opt;
        crc64File({filePath, start, end}, (err, hash) => {
            console.log(filePath, start, end, hash);
            parentPort.postMessage(hash);
        });
    });
}

// 本地调试，主要线程
else if (!module.parent) {
    var time0 = Date.now();
    var filePath = '1g.zip';
    console.log('MAX_THREAD:', MAX_THREAD);
    file_crc64(filePath, (err, hash) => {
        console.log(filePath, hash, (Date.now() - time0) / 1000 + 's');
    });
    // 1g.zip 3534425600523290380n 0.967s
    // 10g.zip 4594563571287695306n 6.13s
}
