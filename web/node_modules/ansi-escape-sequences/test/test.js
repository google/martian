var test = require("tape");
var ansi = require("../");

test("format", function(t){
    t.equal(ansi.format("clive", ["red", "underline"]), "\u001b[31;4mclive\u001b[0m");
    t.end();
});

test("inline format", function(t){
    t.equal(ansi.format("before [red underline]{clive} after"), "before \u001b[31;4mclive\u001b[0m after");
    t.end();
});
