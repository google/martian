'use strict'

/**
Gets a value for a property path.
@module object-get
@typicalname objectGet
@example
var objectGet = require("object-get")
*/
module.exports = objectGet

/**
Returns the value at the given property.
@param {object} - the input object
@param {string} - the property accessor expression
@returns {*}
@alias module:object-get
*/
function objectGet (object, expression) {
  return expression.trim().split('.').reduce(function (prev, curr) {
    return prev && prev[curr]
  }, object)
}
