var test = require('tape')
var a = require('../')

test('.spliceWhile(array, start, regex)', function (t) {
  var array = [ 'one', 'one', 'two', 'three' ]
  t.deepEqual(a.spliceWhile(array, 0, /one/), [ 'one', 'one' ])
  t.deepEqual(array, [ 'two', 'three' ])
  t.end()
})

test('.spliceWhile(array, start, function)', function (t) {
  function under10 (n) { return n < 10; }
  var array = [ 1, 2, 4, 12 ]
  t.deepEqual(a.spliceWhile(array, 0, under10), [ 1, 2, 4 ])
  t.deepEqual(array, [ 12 ])
  t.end()
})

test('.spliceWhile(array, start, [ primitive ])', function (t) {
  var array = [ 1, 2, 4, 12 ]
  t.deepEqual(a.spliceWhile(array, 0, [ 1, 2 ]), [ 1, 2 ])
  t.deepEqual(array, [ 4, 12 ])
  t.end()
})
