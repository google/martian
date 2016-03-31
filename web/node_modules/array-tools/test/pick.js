var test = require('tape')
var a = require('../')

var f = {
  recordset: [
    { one: 'un', two: 'deux', three: 'trois' },
    { two: 'two', one: 'one' },
    { four: 'quattro' },
    { two: 'zwei' }
  ],
  deep: [
    { one: { one: 1, two: 2 }},
    { one: { one: 1, two: 2 }}
  ]
}

test('.pick(recordset, property)', function (t) {
  t.deepEqual(a.pick(f.recordset, 'one'), [
    { one: 'un' },
    { one: 'one' }
  ])
  t.end()
})

test('.pick(recordset, [ properties ])', function (t) {
  t.deepEqual(a.pick(f.recordset, [ 'one', 'two' ]), [
    { one: 'un', two: 'deux' },
    { two: 'two', one: 'one' },
    { two: 'zwei' },
  ])
  t.end()
})

test('.pick(recordset, property.property)', function (t) {
  t.deepEqual(a.pick(f.deep, 'one.two'), [
    { two: 2 },
    { two: 2 },
  ])
  t.end()
})
