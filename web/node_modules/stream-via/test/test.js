var test = require("tape");
var via = require("../");
var PassThrough = require("stream").PassThrough;
var bufferEquals = require("buffer-equal");

test("via(func) - utf8", function(t){
    var stream = via(function(data){
        return data + "yeah?";
    });
    
    stream.on("readable", function(){
        var chunk = this.read();
        if (chunk){
            t.strictEqual(chunk, "cliveyeah?");
            t.end();
        }
    });
    stream.setEncoding("utf8");
    stream.end("clive");
});

test("via(func) - buffer", function(t){
    var stream = via(function(data){
        return Buffer.concat([ data, Buffer([ 2 ])]);
    });
    
    stream.on("readable", function(){
        var chunk = this.read();
        if (chunk){
            t.ok(bufferEquals(chunk, Buffer([ 1, 2 ])));
            t.end();
        }
    });

    stream.end(Buffer([ 1 ]));
});
