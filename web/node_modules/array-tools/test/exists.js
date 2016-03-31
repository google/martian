var test = require('tape')
var a = require('../')

var f = {
  recordset: [
    { result: false, number: 1 },
    { result: false, number: 2 }
  ],
  arr: [ 1, 2, 'three' ]
}

test('.exists(recordset, query)', function (t) {
  t.equal(a.exists(f.recordset, { result: true }), false)
  t.equal(a.exists(f.recordset, { result: false }), true)
  t.equal(a.exists(f.recordset, { result: false, number: 3 }), false)
  t.equal(a.exists(f.recordset, { result: false, number: 2 }), true)
  t.end()
})

test('.exists(array, primitive)', function (t) {
  t.equal(a.exists(f.arr, 0), false)
  t.equal(a.exists(f.arr, 1), true)
  t.equal(a.exists(f.arr, '1'), false)
  t.equal(a.exists(f.arr, 'three'), true)
  t.end()
})
