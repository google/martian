var test = require("tape");
var streamConnect = require("../");
var PassThrough = require("stream").PassThrough;

test("does not pass through pass2", function(t){
    t.plan(1);
    var pass1 = PassThrough();
    pass1.on("data", function(data){
        t.strictEqual(data.toString(), "testing");
    });

    var pass2 = PassThrough();
    pass2.on("data", function(data){
        t.fail("should not fire");
    });

    pass1.end("testing");
});

test("when connected, it does pass through pass2", function(t){
    t.plan(2);
    var pass1 = PassThrough();
    pass1.on("data", function(data){
        t.strictEqual(data.toString(), "testing");
    });

    var pass2 = PassThrough();
    pass2.on("data", function(data){
        t.strictEqual(data.toString(), "testing");
    });
    
    var connected = streamConnect(pass1, pass2);
    connected.end("testing");
});

