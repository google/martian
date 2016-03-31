"use strict";
var Transform = require("stream").Transform;

/**
Create a pipeline of connected streams.

@module stream-connect
@example 
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
*/
module.exports = connect;

/**
Connects two duplex streams together.

@param {external:Duplex} - source stream
@param {external:Duplex} - dest stream, to be connected to
@return {external:Transform}
@alias module:stream-connect
*/
function connect(one, two){
    var transform = new Transform({ objectMode: true });
    one.pipe(two);

    one.on("error", function(err){
        transform.emit("error", err);
    });
    two.on("error", function(err){
        transform.emit("error", err);
    });

    transform._transform = function(chunk, enc, done){
        one.write(chunk);
        done();
    };
    transform._flush = function(){
        one.end();
    };
    two.on("readable", function(){
        transform.push(this.read());
    });
    two.on("end", function(){
        transform.push(null);
    });

    return transform;
}

/**
@external Transform
@see https://nodejs.org/api/stream.html#stream_class_stream_transform
*/
/**
@external Duplex
@see https://nodejs.org/api/stream.html#stream_class_stream_duplex
*/
