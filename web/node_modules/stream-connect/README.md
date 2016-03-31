[![view on npm](http://img.shields.io/npm/v/stream-connect.svg)](https://www.npmjs.org/package/stream-connect)
[![npm module downloads per month](http://img.shields.io/npm/dm/stream-connect.svg)](https://www.npmjs.org/package/stream-connect)
[![Build Status](https://travis-ci.org/75lb/stream-connect.svg?branch=master)](https://travis-ci.org/75lb/stream-connect)
[![Dependency Status](https://david-dm.org/75lb/stream-connect.svg)](https://david-dm.org/75lb/stream-connect)

<a name="module_stream-connect"></a>
## stream-connect
Create a pipeline of connected streams.

**Example**  
```js
> streamConnect = require("stream-connect")
> PassThrough = require("stream").PassThrough

> pass1 = PassThrough()
> pass1.setEncoding("utf8")
> pass1.on("data", console.log.bind(console, "pass1"))

> pass2 = PassThrough()
> pass2.setEncoding("utf8")
> pass2.on("data", console.log.bind(console, "pass2"))

> pass1.write("testing")
pass1 testing

> connected = streamConnect(pass1, pass2)
> connected.write("testing")
pass1 testing
pass2 testing
```
<a name="exp_module_stream-connect--connect"></a>
### connect(one, two) ⇒ <code>[Transform](https://nodejs.org/api/stream.html#stream_class_stream_transform)</code> ⏏
Connects two duplex streams together.

**Kind**: Exported function  

| Param | Type | Description |
| --- | --- | --- |
| one | <code>[Duplex](https://nodejs.org/api/stream.html#stream_class_stream_duplex)</code> | source stream |
| two | <code>[Duplex](https://nodejs.org/api/stream.html#stream_class_stream_duplex)</code> | dest stream, to be connected to |


* * *

&copy; 2015 Lloyd Brookes \<75pound@gmail.com\>. Documented by [jsdoc-to-markdown](https://github.com/jsdoc2md/jsdoc-to-markdown).
