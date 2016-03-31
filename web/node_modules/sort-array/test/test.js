var test = require("tape");
var sortBy = require("../");

test("sortBy", function(t){
    var fixture = [
        { a: 4, b: 1, c: 1},
        { a: 4, b: 3, c: 1},
        { a: 2, b: 2, c: 3},
        { a: 2, b: 2, c: 2},
        { a: 1, b: 3, c: 4},
        { a: 1, b: 1, c: 4},
        { a: 1, b: 2, c: 4},
        { a: 3, b: 3, c: 3},
        { a: 4, b: 3, c: 1}
    ];
    var expected = [
        { a: 1, b: 1, c: 4},
        { a: 1, b: 2, c: 4},
        { a: 1, b: 3, c: 4},
        { a: 2, b: 2, c: 2},
        { a: 2, b: 2, c: 3},
        { a: 3, b: 3, c: 3},
        { a: 4, b: 1, c: 1},
        { a: 4, b: 3, c: 1},
        { a: 4, b: 3, c: 1}
    ];
    t.deepEqual(sortBy(fixture, ["a", "b", "c"]), expected);
    t.end();
});

test("sortBy, with undefined vals", function(t){
    var fixture = [ { a: 1 }, { }, { a: 0 } ];
    var expected = [ { }, { a: 0 }, { a: 1 } ];
    t.deepEqual(sortBy(fixture, "a"), expected);
    t.end();
});

test("sortBy, with undefined vals 2", function(t){
    var fixture = [ { a: "yeah" }, { }, { a: "what" } ];
    var expected = [ { }, { a: "what" }, { a: "yeah" } ];
    t.deepEqual(sortBy(fixture, "a"), expected);
    t.end();
});

test("custom order", function(t){
    var fixture = [{ fruit: "apple" }, { fruit: "orange" }, { fruit: "banana" }, { fruit: "pear" }];
    var expected = [{ fruit: "banana" }, { fruit: "pear" }, { fruit: "apple" }, { fruit: "orange" }];
    var fruitOrder = [ "banana", "pear", "apple", "orange" ];
    t.deepEqual(sortBy(fixture, "fruit", { fruit: fruitOrder }), expected);
    t.end();
});
