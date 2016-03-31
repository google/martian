"use strict";
var test = require("tape");
var arrayify = require("../");

test("arrayify()", function(t){
    t.deepEqual(arrayify(undefined), []);
    t.deepEqual(arrayify(null), [ null ]);
    t.deepEqual(arrayify(0), [ 0 ]);
    t.deepEqual(arrayify([ 1, 2 ]), [ 1, 2 ]);
    
    function func(){
        t.deepEqual(arrayify(arguments), [ 1, 2, 3 ]);
    }
    func(1, 2, 3);

    t.end();
});
