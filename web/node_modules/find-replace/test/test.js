var test = require("tape");
var findReplace = require("../");

function fixture(){
    return [ 1, 2, 3, 4, 2 ];
}
function argv(){
    return [ "--one", "1", "-abc", "three" ];
}

test("find primitive, replace with primitive", function(t){
    t.deepEqual(
        findReplace(fixture(), 2, "two"), 
        [ 1, "two", 3, 4, "two" ]
    );
    t.end();
});

test("find primitive, replace with array", function(t){
    t.deepEqual(
        findReplace(fixture(), 2, [ "two", "zwei" ]), 
        [ 1, [ "two", "zwei" ], 3, 4, [ "two", "zwei" ] ]
    );
    t.end();
});

test("find primitive, replace with several primitives", function(t){
    t.deepEqual(
        findReplace(fixture(), 2, "two", "zwei"), 
        [ 1, "two", "zwei", 3, 4, "two", "zwei" ]
    );
    t.end();
});

test("getopt example", function(t){
    t.deepEqual(
        findReplace(argv(), /^-(\w{2,})$/, function(match){
            return [ "-a", "-b", "-c" ];
        }), 
        [ "--one", "1", "-a", "-b", "-c", "three" ]
    );
    t.end();
});

test("getopt example", function(t){
    t.deepEqual(
        findReplace(argv(), /^-(\w{2,})$/, "bread", "milk"), 
        [ "--one", "1", "bread", "milk", "three" ]
    );
    t.end();
});
