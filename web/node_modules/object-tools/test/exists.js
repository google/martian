var test = require('tape')
var o = require('../')

function TestClass () {
  this.one = 1
}

var testClass = new TestClass()

var fixture = {
  result: 'clive',
  hater: true,
  colour: 'red-ish',
  deep: {
    name: 'Zhana',
    favourite: {
      colour: [ 'white', 'red' ]
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
}

test('.exists(obj, { property: primative })', function (t) {
  t.strictEqual(o.exists(fixture, { result: 'clive' }), true)
  t.strictEqual(o.exists(fixture, { hater: true }), true)
  t.strictEqual(o.exists(fixture, { result: 'clive', hater: true }), true)
  t.strictEqual(o.exists(fixture, { ibe: true }), false)
  t.end()
})

test('.exists(obj, { !property: primative })', function (t) {
  t.strictEqual(o.exists(fixture, { '!result': 'clive' }), false)
  t.strictEqual(o.exists(fixture, { '!result': 'ian' }), true)
  t.strictEqual(o.exists(fixture, { '!result': 'ian', '!hater': false }), true)
  t.end()
})

test('.exists(obj, { property: primative[] })', function (t) {
  t.strictEqual(o.exists(fixture, { arr: [ 1, 2, 3 ] }), true)
  t.strictEqual(o.exists(fixture, { arr: [ /1/ ] }), true)
  t.strictEqual(o.exists(fixture, { arr: [ /4/ ] }), false)
  t.strictEqual(o.exists(fixture, { colour: [ 1, 2, 3 ] }), false, 'querying a string with array')
  t.strictEqual(o.exists(fixture, { undefinedProperty: [ 1, 2, 3 ] }), false, 'querying undefined property')
  t.end()
})

test('.exists(obj, { property: { property: primative[] } })', function (t) {
  t.strictEqual(o.exists(fixture, { deep: { arr: [ 1, 2 ] } }), true)
  t.strictEqual(o.exists(fixture, { deep: { arr: [ 3, 4 ] } }), true)
  t.strictEqual(o.exists(fixture, { deep: { favourite: { colour: [ 'white', 'red' ] } } }), true)
  t.end()
})

test('.exists(obj, { property: undefined, property: regex })', function (t) {
  t.strictEqual(o.exists(fixture.deep, { undefinedProperty: undefined, name: /.+/ }), true)
  t.end()
})

test('.exists(obj, { property: /regex/ })', function (t) {
  t.strictEqual(o.exists(fixture, { colour: /red/ }), true)
  t.strictEqual(o.exists(fixture, { colour: /black/ }), false)
  t.strictEqual(o.exists(fixture, { colour: /RED/i }), true)
  t.strictEqual(o.exists(fixture, { colour: /.+/ }), true)
  t.strictEqual(o.exists(fixture, { undefinedProperty: /.+/ }), false, 'testing undefined val')
  t.strictEqual(o.exists(fixture, { deep: /.+/ }), false, 'testing an object val')
  t.strictEqual(o.exists(fixture, { nullVal: /.+/ }), false, 'testing a null val')
  t.strictEqual(o.exists(fixture, { boolTrue: /true/ }), true, 'testing a boolean val')
  t.strictEqual(o.exists(fixture, { boolTrue: /addf/ }), false, 'testing a boolean val')
  t.end()
})

test('.exists(obj, { !property: /regex/ })', function (t) {
  t.strictEqual(o.exists(fixture, { '!colour': /red/ }), false)
  t.strictEqual(o.exists(fixture, { '!colour': /black/ }), true)
  t.strictEqual(o.exists(fixture, { '!colour': /blue/ }), true)
  t.end()
})

test('.exists(obj, { property: function })', function (t) {
  t.strictEqual(o.exists(fixture, { number: function (n) { return n < 4 } }), false, '< 4')
  t.strictEqual(o.exists(fixture, { number: function (n) { return n < 10 } }), true, '< 10')
  t.end()
})

test('.exists(obj, { !property: function })', function (t) {
  t.strictEqual(o.exists(fixture, { '!number': function (n) { return n < 10 } }), false, '< 10')
  t.end()
})

test('.exists(obj, { property: object })', function (t) {
  t.strictEqual(o.exists(fixture, { testClass: { one: 1 } }), true, 'querying a plain object')
  t.strictEqual(o.exists(fixture, { testClass: testClass }), true, 'querying an object instance')
  t.end()
})

test('.exists(obj, { +property: primitive })', function (t) {
  t.strictEqual(o.exists(fixture, { arr: 1 }), false)
  t.strictEqual(o.exists(fixture, { '+arr': 1 }), true)
  t.end()
})

test('.exists(obj, { property. { +property: query } })', function (t) {
  t.strictEqual(o.exists(fixture, { deep: { favourite: { '+colour': 'red' } } }), true)
  t.strictEqual(o.exists(fixture, { deep: { favourite: { '+colour': /red/ } } }), true)
  t.strictEqual(o.exists(fixture, { deep: { favourite: { '+colour': function (c) {
    return c === 'red'
  } } } }), true)
  t.strictEqual(o.exists(fixture, { deep: { favourite: { '+colour': /green/ } } }), false)
  t.end()
})

test('.exists(obj, { +property: query })', function (t) {
  t.strictEqual(o.exists(fixture, { arrObjects: { number: 1 } }), false)
  t.strictEqual(o.exists(fixture, { '+arrObjects': { number: 1 } }), true)
  t.end()
})

test('object deep exists, summary', function (t) {
  var query = {
    one: {
      one: {
        three: 'three',
        '!four': 'four'
      },
      two: {
        one: {
          one: 'one'
        },
        '!two': undefined,
        '!three': [ { '!one': { '!one': '110' } } ]
      }
    }
  }

  var obj1 = {
    one: {
      one: {
        one: 'one',
        two: 'two',
        three: 'three'
      },
      two: {
        one: {
          one: 'one'
        },
        two: 2
      }
    }
  }

  var obj2 = {
    one: {
      one: {
        one: 'one',
        two: 'two'
      },
      two: {
        one: {
          one: 'one'
        },
        two: 2
      }
    }
  }

  var obj3 = {
    one: {
      one: {
        one: 'one',
        two: 'two',
        three: 'three'
      },
      two: {
        one: {
          one: 'one'
        },
        two: 2,
        three: [
          { one: { one: '100' } },
          { one: { one: '110' } }
        ]
      }
    }
  }

  var obj4 = {
    one: {
      one: {
        one: 'one',
        two: 'two',
        three: 'three'
      },
      two: {
        one: {
          one: 'one'
        },
        two: 2,
        three: [
          { one: { one: '100' } }
        ]
      }
    }
  }

  t.strictEqual(o.exists(obj1, query), true, 'true obj1')
  t.strictEqual(o.exists(obj2, query), false, 'false obj2')
  t.strictEqual(o.exists(obj3, query), false, 'false in obj3')
  t.strictEqual(o.exists(obj4, query), true, 'true in obj4')
  t.end()
})
