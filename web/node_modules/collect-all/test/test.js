var test = require("tape");
var collect = require("../");
var PassThrough = require("stream").PassThrough;

test("collect", function(t){
    var stream = collect();
    stream.on("readable", function(){
        var chunk = this.read();
        if (chunk){
            t.strictEqual(chunk, "onetwo");
            t.end();
        }
    });
    stream.setEncoding("utf8");
    stream.write("one");
    stream.write("two");
    stream.end();
});

test(".collect(options.through)", function(t){
    var stream = collect({
        through: function(data){
            return data + "yeah?";
        }
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

test("passing decodeStrings etc. ");
