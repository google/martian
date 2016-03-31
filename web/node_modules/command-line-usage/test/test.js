var test = require('tape')
var getUsage = require('../')

test('getUsage(definitions, options)', function (t) {
  var definitions = [
    {
      name: 'help', description: 'Display this usage guide.',
      alias: 'h', type: Boolean, group: 'one'
    },
    {
      name: 'src', description: 'The input files to process',
      type: String, multiple: true, defaultOption: true, group: 'one'
    },
    {
      name: 'timeout', description: 'Timeout value in ms',
      alias: 't', type: Number
    }
  ]

  var options = {
    title: 'a typical app',
    description: 'Generates something very important.'
  }

  var result = getUsage(definitions, options)
  t.ok(/a typical app/.test(result))
  t.end()
})

test('getUsage.optionList()', function (t) {
  var definitions = [
    { name: 'one', description: 'one', group: 'one' },
    { name: 'two', description: 'two', group: 'one' },
    { name: 'three', description: 'three' }
  ]

  t.deepEqual(getUsage.optionList(definitions), [
    '  \x1b[1m--one\x1b[0m      one   ',
    '  \x1b[1m--two\x1b[0m      two   ',
    '  \x1b[1m--three\x1b[0m    three '
  ])
  t.deepEqual(getUsage.optionList(definitions, 'one'), [
    '  \x1b[1m--one\x1b[0m    one ',
    '  \x1b[1m--two\x1b[0m    two '
  ])
  t.deepEqual(getUsage.optionList(definitions, '_none'), [
    '  \x1b[1m--three\x1b[0m    three '
  ])
  t.end()
})
