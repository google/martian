var test = require('tape')
var a = require('../')

var f = {
  num: [ 1, 2, 3, 4 ],
  recordset: [
    { n: 1 }, { n: 2 }, { n: 3 }, { n: 4 }
  ]
}

test('.without does not return the input array', function (t) {
  var result = a.without(f.num, 2)
  t.notStrictEqual(f.num, result)
  t.end()
})

test('.without(array, primitive)', function (t) {
  t.deepEqual(a.without(f.num, 2), [ 1, 3, 4 ])
  t.end()
})

test('.without(array, regex)', function (t) {
  t.deepEqual(a.without(f.num, /2/), [ 1, 3, 4 ])
  t.end()
})

test('.without(array, function)', function (t) {
  function over1 (val) { return val > 1; }
  function under4 (val) { return val < 4; }
  t.deepEqual(a.without(f.num, over1), [ 1 ])
  t.end()
})

test('.without(array, query)', function (t) {
  t.deepEqual(a.without(f.recordset, { n: 0}), f.recordset)
  t.deepEqual(a.without(f.recordset, { n: 1}), [
    { n: 2 }, { n: 3 }, { n: 4 }
  ])
  t.end()
})

test('.without(array, array)', function (t) {
  t.deepEqual(a.without(f.num, [ 2, 3 ]), [ 1, 4 ])
  t.end()
})
