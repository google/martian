var test = require('tape')
var a = require('../')

return
test('groupBy', function (t) {
  var fixture = [
    { scope: 'inner',    name: 'one' },
    { scope: 'inner',    name: 'two' },
    { scope: 'instance', name: 'three' },
    { scope: 'instance', name: 'four' },
    { scope: 'global',   name: 'five' },
    { scope: 'global',   name: 'six' },
    { scope: 'global',   name: 'seven' },
    { scope: 'static',   name: 'eight' },
    { scope: 'inner',    name: 'nine' }
  ]
  var expected = [
    {
      group: 'inner',
      items: [
        { scope: 'inner',    name: 'one' },
        { scope: 'inner',    name: 'two' }
      ]
    },
    {
      group: 'instance',
      items: [
        { scope: 'instance', name: 'three' },
        { scope: 'instance', name: 'four' }
      ]
    },
    {
      group: 'global',
      items: [
        { scope: 'global',   name: 'five' },
        { scope: 'global',   name: 'six' },
        { scope: 'global',   name: 'seven' }
      ]
    },
    {
      group: 'static',
      items: [
        { scope: 'static',   name: 'eight' }
      ]
    },
    {
      group: 'inner',
      items: [
        { scope: 'inner',    name: 'nine' }
      ]
    }
  ]
  t.deepEqual(a.groupBy(fixture, 'scope'), expected)
  t.end()
})

test('groupBy, level 2', function (t) {
  var fixture = [
    { scope: 'inner', cat: 'one', name: 'one' },
    { scope: 'inner', cat: 'one', name: 'two' },
    { scope: 'inner', cat: 'two', name: 'three' },
    { scope: 'inner', cat: 'two', name: 'four' },
    { scope: 'inner', cat: 'two', name: 'five' },
    { scope: 'inner', name: 'eight' },
    { scope: 'inner', name: 'nine' }
  ]
  var expected = [
    {
      group: 'inner',
      items: [
        {
          group: 'one',
          items: [
            { scope: 'inner', cat: 'one', name: 'one' },
            { scope: 'inner', cat: 'one', name: 'two' }
          ]
        },
        {
          group: 'two',
          items: [
            { scope: 'inner', cat: 'two', name: 'three' },
            { scope: 'inner', cat: 'two', name: 'four' },
            { scope: 'inner', cat: 'two', name: 'five' }
          ]
        },
        { scope: 'inner', name: 'eight' },
        { scope: 'inner', name: 'nine' }
      ]
    }
  ]
  t.deepEqual(a.groupBy(fixture, 'inner', 'cat'), expected)
  t.end()
})

test('alt groupBy', function (t) {
  var fixture = [
    { scope: 'inner',    name: 'one' },
    { scope: 'inner',    name: 'two' },
    { scope: 'instance', name: 'three' },
    { scope: 'instance', name: 'four' },
    { scope: 'global',   name: 'five' },
    { scope: 'global',   name: 'six' },
    { scope: 'global',   name: 'seven' },
    { scope: 'static',   name: 'eight' },
    { scope: 'inner',    name: 'nine' }
  ]
  var expected = [
    { scope: 'inner',    name: 'one',   _group: 'inner' },
    { scope: 'inner',    name: 'two',   _group: 'inner' },
    { scope: 'instance', name: 'three', _group: 'instance' },
    { scope: 'instance', name: 'four',  _group: 'instance' },
    { scope: 'global',   name: 'five',  _group: 'global' },
    { scope: 'global',   name: 'six',   _group: 'global' },
    { scope: 'global',   name: 'seven', _group: 'global' },
    { scope: 'static',   name: 'eight', _group: 'static' },
    { scope: 'inner',    name: 'nine',  _group: 'inner' }
  ]
  t.deepEqual(a.groupBy(fixture, 'scope'), expected)
  t.end()
})

test('alt groupBy, level 2', function (t) {
  var fixture = [
    { scope: 'inner', cat: 'one', name: 'one' },
    { scope: 'inner', cat: 'one', name: 'two' },
    { scope: 'inner', cat: 'two', name: 'three' },
    { scope: 'inner', cat: 'two', name: 'four' },
    { scope: 'inner', cat: 'two', name: 'five' },
    { name: 'constructor' },
    { name: 'no scope', cat: 'three' },
    { scope: 'inner', name: 'eight' },
    { scope: 'inner', name: 'nine' }
  ]
  var expected = [
    { name: 'constructor', _group: '' },
    { name: 'no scope', cat: 'three', _group: '¦three' },
    { scope: 'inner', name: 'eight', _group: 'inner' },
    { scope: 'inner', name: 'nine', _group: 'inner' },
    { scope: 'inner', cat: 'one', name: 'one', _group: 'inner¦one' },
    { scope: 'inner', cat: 'one', name: 'two', _group: 'inner¦one' },
    { scope: 'inner', cat: 'two', name: 'three', _group: 'inner¦two' },
    { scope: 'inner', cat: 'two', name: 'four', _group: 'inner¦two' },
    { scope: 'inner', cat: 'two', name: 'five', _group: 'inner¦two' }
  ]
  t.deepEqual(a.groupBy(fixture, 'inner', 'cat'), expected)
  t.end()
})

test('alt groupBy', function (t) {
  var fixture = [
    { scope: 'inner',    name: 'one' },
    { scope: 'inner',    name: 'two' },
    { scope: 'instance', name: 'three' },
    { scope: 'instance', name: 'four' },
    { scope: 'global',   name: 'five' },
    { scope: 'global',   name: 'six' },
    { scope: 'global',   name: 'seven' },
    { scope: 'static',   name: 'eight' },
    { scope: 'inner',    name: 'nine' }
  ]
  var expected = [
    {
      group: 'inner',
      items: [
        { scope: 'inner',    name: 'one' },
        { scope: 'inner',    name: 'two' },
      ]
    },
    {
      group: 'instance',
      items: [
        { scope: 'instance', name: 'three' },
        { scope: 'instance', name: 'four' }
      ]
    },
    {
      group: 'global',
      items: [
        { scope: 'global',   name: 'five' },
        { scope: 'global',   name: 'six' },
        { scope: 'global',   name: 'seven' }
      ]
    },
    {
      group: 'static',
      items: [
        { scope: 'static',   name: 'eight' },
      ]
    },
    {
      group: 'inner',
      items: [
        { scope: 'inner',    name: 'nine' }
      ]
    }
  ]
  t.deepEqual(a.groupBy(fixture, 'scope'), expected)
  t.end()
})

test('alt groupBy, level 2', function (t) {
  var fixture = [
    { scope: 'inner', cat: 'one', name: 'one' },
    { scope: 'inner', cat: 'one', name: 'two' },
    { scope: 'inner', cat: 'two', name: 'three' },
    { scope: 'inner', cat: 'two', name: 'four' },
    { scope: 'inner', cat: 'two', name: 'five' },
    { scope: 'inner', name: 'eight' },
    { scope: 'inner', name: 'nine' }
  ]
  var expected = [
    {
      group: 'inner',
      items: [
        { scope: 'inner', name: 'eight' },
        { scope: 'inner', name: 'nine' }
      ]
    },
    {
      group: 'inner¦one',
      items: [
        { scope: 'inner', cat: 'one', name: 'one' },
        { scope: 'inner', cat: 'one', name: 'two' }
      ]
    },
    {
      group: 'inner¦two',
      items: [
        { scope: 'inner', cat: 'two', name: 'three' },
        { scope: 'inner', cat: 'two', name: 'four' },
        { scope: 'inner', cat: 'two', name: 'five' }
      ]
    }
  ]
  t.deepEqual(a.groupBy(fixture, 'inner', 'cat'), expected)
  t.end()
})
