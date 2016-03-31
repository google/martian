var test = require("tape");
var testValue = require("../");

function TestClass(){
    this.one = 1;
}

var testClass = new TestClass();

var fixture = {
    result: "clive",
    hater: true,
    colour: "red-ish",
    deep: {
        name: "Zhana",
        favourite: {
            colour: [ "white", "red" ]
        },
        arr: [ 1, 2, 3 ]
    },
    nullVal: null,
    boolTrue: true,
    number: 5,
    testClass: testClass,
    arr: [ 1, 2, 3 ],
    arrObjects: [
        { number: 1 },
        { number: 2 }
    ]
};

test(".exists(obj, { property: primative })", function(t){
    t.strictEqual(testValue(fixture, { result: "clive" }), true);
    t.strictEqual(testValue(fixture, { hater: true }), true);
    t.strictEqual(testValue(fixture, { result: "clive", hater: true }), true);
    t.strictEqual(testValue(fixture, { ibe: true }), false);
    t.end();
});

test(".exists(obj, { !property: primative })", function(t){
    t.strictEqual(testValue(fixture, { "!result": "clive" }), false);
    t.strictEqual(testValue(fixture, { "!result": "ian" }), true);
    t.strictEqual(testValue(fixture, { "!result": "ian", "!hater": false }), true);
    t.end();
});

test(".exists(obj, { property: primative[] })", function(t){
    t.strictEqual(testValue(fixture, { arr: [ 1, 2, 3 ] }), true);
    t.strictEqual(testValue(fixture, { arr: [ /1/ ] }), true);
    t.strictEqual(testValue(fixture, { arr: [ /4/ ] }), false);
    t.strictEqual(testValue(fixture, { colour: [ 1, 2, 3 ] }), false, "querying a string with array");
    t.strictEqual(testValue(fixture, { undefinedProperty: [ 1, 2, 3 ] }), false, "querying undefined property");
    t.strictEqual(testValue(fixture, { undefinedProperty: [ undefined ] }), true);
    t.strictEqual(testValue(fixture, { undefinedProperty: [ null ] }), false);
    t.end();
});

test(".exists(obj, { property: { property: primative[] } })", function(t){
    t.strictEqual(testValue(fixture, { deep: { arr: [ 1, 2 ] } }), true);
    t.strictEqual(testValue(fixture, { deep: { arr: [ 3, 4 ] } }), true);
    t.strictEqual(testValue(fixture, { deep: { favourite: { colour: [ "white", "red" ] } } }), true);
    t.end();
});

test(".exists(obj, { property: undefined, property: regex })", function(t){
    t.strictEqual(testValue(fixture.deep, { undefinedProperty: undefined, name: /.+/ }), true);
    t.end();
});

test(".exists(obj, { property: /regex/ })", function(t){
    t.strictEqual(testValue(fixture, { colour: /red/ }), true);
    t.strictEqual(testValue(fixture, { colour: /black/ }), false);
    t.strictEqual(testValue(fixture, { colour: /RED/i }), true);
    t.strictEqual(testValue(fixture, { colour: /.+/ }), true);
    t.strictEqual(testValue(fixture, { undefinedProperty: /.+/ }), false, "testing undefined val");
    t.strictEqual(testValue(fixture, { deep: /.+/ }), false, "testing an object val");
    t.strictEqual(testValue(fixture, { nullVal: /.+/ }), false, "testing a null val");
    t.strictEqual(testValue(fixture, { boolTrue: /true/ }), true, "testing a boolean val");
    t.strictEqual(testValue(fixture, { boolTrue: /addf/ }), false, "testing a boolean val");
    t.end();
});

test(".exists(obj, { !property: /regex/ })", function(t){
    t.strictEqual(testValue(fixture, { "!colour": /red/ }), false);
    t.strictEqual(testValue(fixture, { "!colour": /black/ }), true);
    t.strictEqual(testValue(fixture, { "!colour": /blue/ }), true);
    t.end();
});

test(".exists(obj, { property: function })", function(t){
    t.strictEqual(testValue(fixture, { number: function(n){ return n < 4; }}), false, "< 4");
    t.strictEqual(testValue(fixture, { number: function(n){ return n < 10; }}), true, "< 10");
    t.end();
});

test(".exists(obj, { !property: function })", function(t){
    t.strictEqual(testValue(fixture, { "!number": function(n){ return n < 10; }}), false, "< 10");
    t.end();
});

test(".exists(obj, { property: object })", function(t){
    t.strictEqual(testValue(fixture, { testClass: { one: 1 } }), true, "querying a plain object");
    t.strictEqual(testValue(fixture, { testClass: testClass }), true, "querying an object instance");
    t.end();
});


test(".exists(obj, { +property: primitive })", function(t){
    t.strictEqual(testValue(fixture, { arr: 1 }), false);
    t.strictEqual(testValue(fixture, { "+arr": 1 }), true);
    t.end();
});

test(".exists(obj, { property. { +property: query } })", function(t){
    t.strictEqual(testValue(fixture, { deep: { favourite: { "+colour": "red" } } }), true);
    t.strictEqual(testValue(fixture, { deep: { favourite: { "+colour": /red/ } } }), true);
    t.strictEqual(testValue(fixture, { deep: { favourite: { "+colour": function(c){ 
        return c === "red"; 
    } } } }), true);
    t.strictEqual(testValue(fixture, { deep: { favourite: { "+colour": /green/ } } }), false);
    t.end();
});

test(".exists(obj, { +property: query })", function(t){
    t.strictEqual(testValue(fixture, { arrObjects: { number: 1 } }), false);
    t.strictEqual(testValue(fixture, { "+arrObjects": { number: 1 } }), true);
    t.end();
});

test("object deep exists, summary", function(t){
    var query = {
        one: {
            one: {
                three: "three",
                "!four": "four"
            },
            two: {
                one: {
                    one: "one"
                },
                "!two": undefined,
                "!three": [ { "!one": { "!one": "110" } } ]
            }
        }
    };

    var obj1 = {
        one: {
            one: {
                one: "one",
                two: "two",
                three: "three"
            },
            two: {
                one: {
                    one: "one"
                },
                two: 2
            }
        }
    };

    var obj2 = {
        one: {
            one: {
                one: "one",
                two: "two"
            },
            two: {
                one: {
                    one: "one"
                },
                two: 2
            }
        }
    };

    var obj3 = {
        one: {
            one: {
                one: "one",
                two: "two",
                three: "three"
            },
            two: {
                one: {
                    one: "one"
                },
                two: 2,
                three: [
                    { one: { one: "100" } },
                    { one: { one: "110" } }
                ]
            }
        }
    };

    var obj4 = {
        one: {
            one: {
                one: "one",
                two: "two",
                three: "three"
            },
            two: {
                one: {
                    one: "one"
                },
                two: 2,
                three: [
                    { one: { one: "100" } }
                ]
            }
        }
    };

    t.strictEqual(testValue(obj1, query), true, "true obj1");
    t.strictEqual(testValue(obj2, query), false, "false obj2");
    t.strictEqual(testValue(obj3, query), false, "false in obj3");
    t.strictEqual(testValue(obj4, query), true, "true in obj4");
    t.end();
});
