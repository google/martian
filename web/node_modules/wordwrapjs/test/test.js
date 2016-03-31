var test = require('tape')
var wrap = require('../')

var bars = "I'm rapping. I'm rapping. I'm rap rap rapping. I'm rap rap rap rap rappity rapping."

test('simple', function (t) {
  t.strictEqual(
    wrap(bars),
    "I'm rapping. I'm rapping. I'm\nrap rap rapping. I'm rap rap\nrap rap rappity rapping."
  )
  t.end()
})

test('width', function (t) {
  t.strictEqual(
    wrap(bars, { width: 3 }),
    "I'm\nrapping.\nI'm\nrapping.\nI'm\nrap\nrap\nrapping.\nI'm\nrap\nrap\nrap\nrap\nrappity\nrapping."
  )
  t.end()
})

test('ignore', function (t) {
  t.strictEqual(
    wrap(bars, { ignore: "I'm" }),
    "I'm rapping. I'm rapping. I'm rap rap\nrapping. I'm rap rap rap rap\nrappity rapping."
  )
  t.end()
})

test('wrap.lines', function (t) {
  t.deepEqual(
    wrap.lines(bars),
    [ "I'm rapping. I'm rapping. I'm",
      "rap rap rapping. I'm rap rap",
      'rap rap rappity rapping.' ]
  )
  t.end()
})

test('wrap.lines, width', function (t) {
  t.deepEqual(
    wrap.lines(bars, { width: 3 }),
    [ "I'm",
      'rapping.',
      "I'm",
      'rapping.',
      "I'm",
      'rap',
      'rap',
      'rapping.',
      "I'm",
      'rap',
      'rap',
      'rap',
      'rap',
      'rappity',
      'rapping.' ]
  )
  t.end()
})

test('wrap.lines, width smaller than content width', function (t) {
  t.deepEqual(
    wrap.lines('4444', { width: 3 }),
    [ '4444' ]
  )
  t.deepEqual(
    wrap.lines('onetwothreefour fivesixseveneight', { width: 7 }),
    [ 'onetwothreefour', 'fivesixseveneight' ]
  )

  t.end()
})

test('wrap.lines, break', function (t) {
  t.deepEqual(
    wrap.lines('onetwothreefour', { width: 7, break: true }),
    [ 'onetwot', 'hreefou', 'r' ]
  )
  t.deepEqual(
    wrap.lines('\u001b[4m--------\u001b[0m', { width: 10, break: true, ignore: /\u001b.*?m/g }),
    [ '\u001b[4m--------\u001b[0m' ]
  )
  t.deepEqual(
    wrap.lines(
      'onetwothreefour fivesixseveneight',
      { width: 7, break: true }
    ),
    [ 'onetwot', 'hreefou', 'r', 'fivesix', 'sevenei', 'ght' ]
  )

  t.end()
})

test('wrap.lines(text): respect existing linebreaks', function (t) {
  t.deepEqual(
    wrap.lines('one\ntwo three four', { width: 8 }),
    [ 'one', 'two', 'three', 'four' ]
  )

  t.deepEqual(
    wrap.lines('one \n \n two three four', { width: 8 }),
    [ 'one', '', 'two', 'three', 'four' ]
  )

  t.deepEqual(
    wrap.lines('one\rtwo three four', { width: 8 }),
    [ 'one', 'two', 'three', 'four' ]
  )

  t.deepEqual(
    wrap.lines('one\r\ntwo three four', { width: 8 }),
    [ 'one', 'two', 'three', 'four' ]
  )

  t.end()
})

test('wrap.lines(text): multilingual', function (t) {
  t.deepEqual(
    wrap.lines('Può parlare più lentamente?', { width: 10 }),
    [ 'Può', 'parlare', 'più', 'lentamente?' ]
  )

  t.deepEqual(
    wrap.lines('один два три', { width: 4 }),
    [ 'один', 'два', 'три' ]
  )

  t.end()
})

test('wrap hyphenated words', function (t) {
  t.deepEqual(
    wrap.lines('ones-and-twos', { width: 5 }),
    [ 'ones-', 'and-', 'twos' ]
  )

  t.deepEqual(
    wrap.lines('ones-and-twos', { width: 10 }),
    [ 'ones-and-', 'twos' ]
  )

  t.deepEqual(
    wrap.lines('--------', { width: 5 }),
    [ '--------' ]
  )

  t.deepEqual(
    wrap.lines('--one --fifteen', { width: 5 }),
    [ '--one', '--fifteen' ]
  )

  t.deepEqual(
    wrap.lines('one-two', { width: 10 }),
    [ 'one-two' ]
  )

  t.deepEqual(
    wrap.lines('ansi-escape-sequences', { width: 22 }),
    [ 'ansi-escape-sequences' ]
  )

  t.end()
})

test('isWrappable(input)', function(t){
  t.strictEqual(wrap.isWrappable('one two'), true)
  t.strictEqual(wrap.isWrappable('one-two'), true)
  t.strictEqual(wrap.isWrappable('one\ntwo'), true)
  t.end()
})
