var test = require('tape')
var cliArgs = require('../')

var optionDefinitions = [
  { name: 'one', type: Boolean }
]

test('usage: simple', function (t) {
  var cli = cliArgs(optionDefinitions)
  var usage = cli.getUsage({ title: 'test' })
  t.ok(/test/.test(usage), 'title present')
  t.end()
})
