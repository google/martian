var test = require('tape')
var a = require('../')

var f = {
  arr: [ 1, 1, 2, 3, 4 ],
  recordset: [
    { b: false, n: 1 },
    { b: false, n: 2 }
  ]
}

test('.where(recordset, query)', function (t) {
  t.deepEqual(a.where(f.recordset, { b: true }), [])
  t.deepEqual(a.where(f.recordset, { b: false }), [
    { b: false, n: 1 },
    { b: false, n: 2 }
  ])
  t.deepEqual(a.where(f.recordset, { b: false, n: 3 }), [])
  t.deepEqual(a.where(f.recordset, { b: false, n: 2 }), [
    { b: false, n: 2 }
  ])
  t.end()
})

test('.where(recordset, regex)', function (t) {
  t.deepEqual(a.where(f.recordset, { n: /1/ }), [ { b: false, n: 1 } ])
  t.deepEqual(a.where(f.recordset, { x: undefined, n: /.+/ }), [
    { b: false, n: 1 },
    { b: false, n: 2 }
  ])
  t.end()
})

test('.where(array, primitive)', function (t) {
  t.deepEqual(a.where(f.arr, 1), [ 1, 1 ])
  t.deepEqual(a.where(f.arr, 2), [ 2 ])
  t.end()
})

test('.where(array, regex)', function (t) {
  t.deepEqual(a.where(f.arr, /1/), [ 1, 1 ])
  t.deepEqual(a.where(f.arr, /2/), [ 2 ])
  t.end()
})

test('.where(array, function)', function (t) {
  function over3 (val) { return val > 3; }
  t.deepEqual(a.where(f.arr, over3), [ 4 ])
  t.end()
})

test('.where(array, array)', function (t) {
  function over3 (val) { return val > 3; }
  t.deepEqual(a.where(f.arr, [ 1, /2/, over3 ]), [ 1, 1, 2, 4 ])
  t.end()
})

test('.where(array, object[])', function (t) {
  t.deepEqual(a.where(f.recordset, [ { n: 1 }, { n: 2 }, { n: 3 } ]), [
    { b: false, n: 1 },
    { b: false, n: 2 }
  ])
  t.end()
})

test('.where deep query', function (t) {
  var arr = [
    { one: { number: 1, letter: 'a' } },
    { one: { number: 2, letter: 'b' } },
    { one: { number: 3, letter: 'b' } }
  ]
  t.deepEqual(a.where(arr, { one: { letter: 'b' } }), [
    { one: { number: 2, letter: 'b' } },
    { one: { number: 3, letter: 'b' } }
  ])
  t.deepEqual(a.where(arr, { one: { number: 2, letter: 'b' } }), [
    { one: { number: 2, letter: 'b' } }
  ])
  t.deepEqual(a.where(arr, { one: { number: 1, letter: 'a' } }), [
    { one: { number: 1, letter: 'a' } }
  ])
  t.end()
})
