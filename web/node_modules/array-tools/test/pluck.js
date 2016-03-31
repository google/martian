var test = require('tape')
var a = require('../')

var fixture = [
  {one: 1, two: 2},
  {two: 'two'},
  {one: 'one', two: 'zwei'},
  {deep: { a: 'yep' }},
  {deep: { a: 'again' }},
  {deep: 'not here' },
  {deep: { b: 'or here' }}
]

test('.pluck(array, property)', function (t) {
  t.deepEqual(a.pluck(fixture, 'one'), [ 1, 'one' ])
  t.deepEqual(a.pluck(fixture, 'two'), [ 2, 'two', 'zwei' ])
  t.end()
})

test('.pluck(array, [property])', function (t) {
  t.deepEqual(a.pluck(fixture, [ 'one', 'two' ]), [ 1, 'two', 'one' ])
  t.end()
})

test('.pluck(array, property.property)', function (t) {
  t.deepEqual(a.pluck(fixture, 'deep.a'), [ 'yep', 'again' ])
  t.end()
})
