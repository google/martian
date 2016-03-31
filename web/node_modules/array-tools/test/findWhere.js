var test = require('tape')
var a = require('../')

test('findWhere', function (t) {
  var arr = [
    { result: false, number: 1 },
    { result: false, number: 2 }
  ]
  t.deepEqual(a.findWhere(arr, { result: true }), undefined)
  t.deepEqual(a.findWhere(arr, { result: false }), { result: false, number: 1 })
  t.deepEqual(a.findWhere(arr, { result: false, number: 3 }), undefined)
  t.deepEqual(a.findWhere(arr, { result: false, number: 2 }), { result: false, number: 2 })
  t.end()
})

test('.findWhere deep query', function (t) {
  var arr = [
    { one: { number: 1, letter: 'a' } },
    { one: { number: 2, letter: 'b' } }
  ]
  t.deepEqual(a.findWhere(arr, { one: { number: 1 } }), { one: { number: 1, letter: 'a' } })
  t.deepEqual(a.findWhere(arr, { one: { number: 2 } }), { one: { number: 2, letter: 'b' } })
  t.deepEqual(a.findWhere(arr, { one: { letter: 'b' } }), { one: { number: 2, letter: 'b' } })
  t.deepEqual(a.findWhere(arr, { one: { number: 3 } }), undefined)
  t.end()
})

test('.findWhere deeper query', function (t) {
  var query
  var arr = [
    {
      name: 'one',
      data: { two: { three: 'four' } }
    },
    {
      name: 'two',
      data: { two: { three: 'five' } }
    }
  ]
  query = {  name: 'one', data: { two: { three: 'four' } } }
  t.deepEqual(a.findWhere(arr, query), {
    name: 'one',
    data: { two: { three: 'four' } }
  })
  query = {  name: 'one' }
  t.deepEqual(a.findWhere(arr, query), {
    name: 'one',
    data: { two: { three: 'four' } }
  })
  query = {  name: 'two' }
  t.deepEqual(a.findWhere(arr, query), {
    name: 'two',
    data: { two: { three: 'five' } }
  })

  query = {  name: 'two', data: { two: { three: 'four' } } }
  t.deepEqual(a.findWhere(arr, query), undefined)

  query = {  name: 'two', data: { two: { three: 'five' } } }
  t.deepEqual(a.findWhere(arr, query), {
    name: 'two',
    data: { two: { three: 'five' } }
  })

  t.end()
})
