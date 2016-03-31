switch (process.env.VERSION) {
  case '5.0':
    console.log('Running 5.0 tests')
    require('./es6-5.0')
    break
  case '4.1':
    console.log('Running 4.1 tests')
    require('./es6-4.1')
    break
  case 'iojs':
    console.log('Running iojs tests')
    require('./es6-iojs')
    break
  case '0.12':
    console.log('Running es5-0.12 tests')
    require('./es5-0.12')
    break
  case '0.10':
    console.log('Running es5-0.10 tests')
    require('./es5-0.10')
    break
}
