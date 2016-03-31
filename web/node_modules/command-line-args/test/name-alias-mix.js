var test = require('tape')
var cliArgs = require('../')

var optionDefinitions = [
  { name: 'one', alias: 'o' },
  { name: 'two', alias: 't' },
  { name: 'three', alias: 'h' },
  { name: 'four', alias: 'f' }
]

test('name-alias-mix: one of each', function (t) {
  var argv = [ '--one', '-t', '--three' ]
  var cli = cliArgs(optionDefinitions)
  var result = cli.parse(argv)
  t.strictEqual(result.one, true)
  t.strictEqual(result.two, true)
  t.strictEqual(result.three, true)
  t.strictEqual(result.four, undefined)
  t.end()
})
