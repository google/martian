var test = require('tape')
var fs = require('fs')
var spawn = require('child_process').spawn

var inputPath = 'test/fixture/array.json'
var outputPath = 'tmp/test.json'

try {
  fs.mkdirSync('tmp')
} catch(err) {
  // dir exists
}

test('cli: stdin check', function (t) {
  t.plan(1)

  var inputFile = fs.openSync(inputPath, 'r')
  var outputFile = fs.openSync(outputPath, 'w')

  var handle = spawn('node', [ 'bin/cli.js', 'pluck', 'one' ], {
    stdio: [ inputFile, outputFile, process.stderr ]
  })
  handle.on('close', function () {
    var result = fs.readFileSync(outputPath, 'utf8')
    if (result) t.deepEqual(JSON.parse(result), [ 1 ])
  })
})
