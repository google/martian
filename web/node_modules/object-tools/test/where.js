var test = require('tape')
var o = require('../')

test('.where(object, function)', function (t) {
  var object = { a: 1, b: 0, c: 2 }
  var result = o.where(object, function (value, key) {
    return value > 0
  })
  t.deepEqual(result, { a: 1, c: 2 })
  t.end()
})

test('.where(object, propertyArray)', function (t) {
  var object = { a: 1, b: 0, c: 2 }
  var result = o.where(object, [ 'b' ])
  t.deepEqual(result, { b: 0 })
  t.end()
})

test('defined', function (t) {
  var object = { a: 1, b: undefined, c: 2 }
  var result = o.where(object, function (value, key) {
    return value !== undefined
  })
  t.deepEqual(result, { a: 1, c: 2 })
  t.end()
})
