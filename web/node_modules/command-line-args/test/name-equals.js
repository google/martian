var test = require('tape')
var cliArgs = require('../')

var optionDefinitions = [
  { name: 'one' },
  { name: 'two' },
  { name: 'three' }
]

test('name-alias-mix: one of each', function (t) {
  var argv = [ '--one=1', '--two', '2', '--three=3' ]
  var cli = cliArgs(optionDefinitions)
  var result = cli.parse(argv)
  t.strictEqual(result.one, '1')
  t.strictEqual(result.two, '2')
  t.strictEqual(result.three, '3')
  t.end()
})
