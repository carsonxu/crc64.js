# CRC64-ECMA182.js

Pure JavaScript implement of CRC64-ECMA182 for Node.js.

> This package can be used as verify [Ali-OSS](https://help.aliyun.com/document_detail/43394.html) file.

## Usage

### Calculate a Buffer

You can calculate the CRC64-ECMA182 value for a Node.js buffer or string:

```js
crc64.crc64(buff[, prev]);
```

+ Parameters:
    + `buff`: the buffer or string to be calculated;
    + \[`prev`]: if exists, `prev` indicates the previous CRC64-ECMA182 value; (**optional**)
+ Returns: the result string indicates a uint64 value that calculated.


```js
const crc64 = require('./crc64');
const ret1 = crc64.crc64("123456789");
const ret2 = crc64.crc64(new Buffer("123456789"));
const ret3 = crc64.crc64("123456789", "0");
const ret4 = crc64.crc64(new Buffer("123456789"), "0");

// ret1 ~ ret2 all equals to:
//
//   '11051210869376104954'
```

### Calculate a file

You can calculate the CRC64-ECMA182 value for a file:

```js
crc64.crc64_file(filename, callback);
```

+ Parameters:
  + `filename`: the file's name that to be calculated;
  + `callback`: the callback function which receives two arguments `err` and `ret`.

```js
crc64.crc64_file(path.join(__dirname, "pic.png"), function(err, ret) {
  console.log(err, ret);

  // a possible result:
  //
  //   undefined 5178350320981835788
});
```

### concat file chunks crc64

You can combine file parts crc64 into a total crc64:

```js
crc64.crc64_concat([{hash1, size1}, {hash2, size2}], callback)
```

+ Parameters:
  + `hashList`: crc64 info list, [{hash1, size1}, {hash2, size2}]
  + `callback`: the callback function which receives two arguments `err` and `ret`.

```js
var list = ['123', '456'];
var crc64List = list.map(str => {
  var size = str.length;
  var hash = crc64.crc64(str);
  console.log(`${str}: ${hash}, 0x${hash.toString(16)}`);
  return {hash, size};
});
var hash = crc64.crc64_concat(crc64List);
console.log(`${list.join('')}: ${hash}, 0x${hash.toString(16)}`);
```

## Contribution

You're welcome to make pull-requests.
