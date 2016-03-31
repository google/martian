var detect = require('feature-detect-es6')

module.exports = detect.class() && detect.arrowFunction() && detect.templateStrings()
  ? require('./lib/column-layout')
  : require('./es5/column-layout')
