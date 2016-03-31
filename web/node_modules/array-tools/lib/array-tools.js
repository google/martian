'use strict'
var objectGet = require('object-get')
var arrayify = require('array-back')
var sortBy = require('sort-array')
var testValue = require('test-value')

/**
@module
@typicalname a
*/
module.exports = ArrayTools

ArrayTools.pluck = pluck
ArrayTools.pick = pick
ArrayTools.arrayify = arrayify
ArrayTools.exists = exists
ArrayTools.without = without
ArrayTools.where = where
ArrayTools.findWhere = findWhere
ArrayTools.unique = unique
ArrayTools.spliceWhile = spliceWhile
ArrayTools.extract = extract
ArrayTools.remove = remove
ArrayTools.flatten = flatten
ArrayTools.sortBy = sortBy
ArrayTools.last = last
ArrayTools.contains = contains

function ArrayTools (input) {
  if (!(this instanceof ArrayTools)) return new ArrayTools(input)
  this._data = input
  this.val = function () {
    var endValue = this._data
    this._data = input
    return endValue
  }
}

/* Array methods which return the chainable */
['filter', 'reverse', 'sort', 'concat', 'slice', 'map'].forEach(function (method) {
  ArrayTools.prototype[method] = function () {
    this._data = Array.prototype[method].apply(this._data, arguments)
    return this
  }
})

/* Array method chain terminators, return a scalar or undefined */
;['join', 'every', 'some', 'forEach'].forEach(function (method) {
  ArrayTools.prototype[method] = function () {
    var endValue = this.val()
    return Array.prototype[method].apply(endValue, arguments)
  }
})

/* array-tools methods which return the chainable */
;['pluck', 'pick', 'arrayify', 'where', 'without', 'unique', 'spliceWhile', 'extract', 'flatten', 'sortBy'].forEach(function (method) {
  ArrayTools.prototype[method] = function () {
    var args = arrayify(arguments)
    args.unshift(this._data)
    this._data = ArrayTools[method].apply(null, args)
    return this
  }
})

/* array-tools method chain terminators, return a scalar or non-array */
;['exists', 'findWhere', 'last', 'remove', 'contains'].forEach(function (method) {
  ArrayTools.prototype[method] = function () {
    var args = arrayify(arguments)
    var endValue = this.val()
    args.unshift(endValue)
    return ArrayTools[method].apply(null, args)
  }
})

/**
Takes any input and guarantees an array back.

- converts array-like objects (e.g. `arguments`) to a real array
- converts `undefined` to an empty array
- converts any another other, singular value (including `null`) into an array containing that value
- ignores input which is already an array

@param any {*} - the input value to convert to an array
@returns {Array}
@category chainable
@method arrayify
@static
@example
> a.arrayify(undefined)
[]

> a.arrayify(null)
[ null ]

> a.arrayify(0)
[ 0 ]

> a.arrayify([ 1, 2 ])
[ 1, 2 ]

> function f(){ return a.arrayify(arguments); }
> f(1,2,3)
[ 1, 2, 3 ]
*/

/**
Deep query an array.

@param {object[]} - the array to query
@param {any | any[]} - one or more queries
@returns {Array}
@category chainable
@static
@example
Say you have a recordset:
```js
> data = [
    { name: "Dana", age: 30 },
    { name: "Yana", age: 20 },
    { name: "Zhana", age: 10 }
]
```

You can return records with properties matching an exact value:
```js
> a.where(data, { age: 10 })
[ { name: 'Zhana', age: 10 } ]
```

or where NOT the value (prefix the property name with `!`)
```js
> a.where(data, { "!age": 10 })
[ { name: 'Dana', age: 30 }, { name: 'Yana', age: 20 } ]
```

match using a function:
```js
> function over10(age){ return age > 10; }
> a.where(data, { age: over10 })
[ { name: 'Dana', age: 30 }, { name: 'Yana', age: 20 } ]
```

match using a regular expression
```js
> a.where(data, { name: /ana/ })
[ { name: 'Dana', age: 30 },
  { name: 'Yana', age: 20 },
  { name: 'Zhana', age: 10 } ]
```

You can query to any arbitrary depth. So with deeper data, like this:
```js
> deepData = [
    { name: "Dana", favourite: { colour: "light red" } },
    { name: "Yana", favourite: { colour: "dark red" } },
    { name: "Zhana", favourite: { colour: [ "white", "red" ] } }
]
```

get records with `favourite.colour` values matching `/red/`
```js
> a.where(deepData, { favourite: { colour: /red/ } })
[ { name: 'Dana', favourite: { colour: 'light red' } },
  { name: 'Yana', favourite: { colour: 'dark red' } } ]
```

if the value you're looking for _maybe_ part of an array, prefix the property name with `+`. Now Zhana is included:
```js
> a.where(deepData, { favourite: { "+colour": /red/ } })
[ { name: 'Dana', favourite: { colour: 'light red' } },
  { name: 'Yana', favourite: { colour: 'dark red' } },
  { name: 'Zhana', favourite: { colour: [ "white", "red" ] } } ]
```

you can combine any of the above by supplying an array of queries. Records will be returned if _any_ of the queries match:
```js
> var nameBeginsWithY = { name: /^Y/ }
> var faveColourIncludesWhite = { favourite: { "+colour": "white" } }

> a.where(deepData, [ nameBeginsWithY, faveColourIncludesWhite ])
[ { name: 'Yana', favourite: { colour: 'dark red' } },
  { name: 'Zhana', favourite: { colour: [ "white", "red" ] } } ]
```
*/
function where (array, query) {
  array = arrayify(array)
  return array.filter(function (item) {
    return testValue(item, query)
  })
}

/**
Returns a new array with the same content as the input minus the specified values. It accepts the same query syntax as {@link module:array-tools.where}.

@param {Array} - the input array
@param {any | any[]} - one, or more queries
@returns {Array}
@category chainable
@example
> a.without([ 1, 2, 3 ], 2)
[ 1, 3 ]

> a.without([ 1, 2, 3 ], [ 2, 3 ])
[ 1 ]

> data = [
    { name: "Dana", age: 30 },
    { name: "Yana", age: 20 },
    { name: "Zhana", age: 10 }
]
> a.without(data, { name: /ana/ })
[]

@alias module:array-tools.without
*/
function without (array, toRemove) {
  toRemove = arrayify(toRemove)
  return array.filter(function (item) {
    return !testValue(item, toRemove)
  })
}

/**
Works in exactly the same way as {@link module:array-tools.where} but returning a boolean indicating whether a matching record exists.

@param {Array} - the array to search
@param {*} - the value to search for
@returns {boolean}
@category not chainable
@static
@example
> data = [
    { name: "Dana", age: 30 },
    { name: "Yana", age: 20 },
    { name: "Zhana", age: 10 }
]

> a.exists(data, { age: 10 })
true
*/
function exists (array, query) {
  return arrayify(array).some(function (item) {
    return testValue(item, query)
  })
}

/**
Returns an array containing each value plucked from the specified property of each object in the input array.

@param recordset {object[]} - The input recordset
@param property {string|string[]} - Property name, or an array of property names. If an array is supplied, the first existing property will be returned.
@returns {Array}
@category chainable
@static
@example
with this data..
```js
> var data = [
    { name: "Pavel", nick: "Pasha" },
    { name: "Richard", nick: "Dick" },
    { name: "Trevor" },
]
```

pluck all the nicknames
```js
> a.pluck(data, "nick")
[ 'Pasha', 'Dick' ]
```

in the case no nickname exists, take the name instead:
```js
> a.pluck(data, [ "nick", "name" ])
[ 'Pasha', 'Dick', 'Trevor' ]
```

the values being plucked can be at any depth:
```js
> var data = [
    { leeds: { leeds: { leeds: "we" } } },
    { leeds: { leeds: { leeds: "are" } } },
    { leeds: { leeds: { leeds: "Leeds" } } }
]

> a.pluck(data, "leeds.leeds.leeds")
[ 'we', 'are', 'Leeds' ]
```
*/
function pluck (recordset, property) {
  recordset = arrayify(recordset)
  var properties = arrayify(property)

  return recordset
    .map(function (record) {
      for (var i = 0; i < properties.length; i++) {
        var propValue = objectGet(record, properties[i])
        if (propValue) return propValue
      }
    })
    .filter(function (record) {
      return typeof record !== 'undefined'
    })
}

/**
return a copy of the input `recordset` containing objects having only the cherry-picked properties
@param recordset {object[]} - the input
@param property {string|string[]} - the properties to include in the result
@return {object[]}
@category chainable
@static
@example
with this data..
```js
> data = [
    { name: "Dana", age: 30 },
    { name: "Yana", age: 20 },
    { name: "Zhana", age: 10 }
]
```

return only the `"name"` field..
```js
> a.pick(data, "name")
[ { name: 'Dana' }, { name: 'Yana' }, { name: 'Zhana' } ]
```

return both the `"name"` and `"age"` fields
```js
> a.pick(data, [ "name", "age" ])
[ { name: 'Dana', age: 30 },
  { name: 'Yana', age: 20 },
  { name: 'Zhana', age: 10 } ]
```

cherry-picks fields at any depth:
```js
> data = [
    { person: { name: "Dana", age: 30 }},
    { person: { name: "Yana", age: 20 }},
    { person: { name: "Zhana", age: 10 }}
]

> a.pick(data, "person.name")
[ { name: 'Dana' }, { name: 'Yana' }, { name: 'Zhana' } ]

> a.pick(data, "person.age")
[ { age: 30 }, { age: 20 }, { age: 10 } ]
```
*/
function pick (recordset, property) {
  recordset = arrayify(recordset)
  var properties = arrayify(property)

  return recordset
    .filter(function (obj) {
      return properties.some(function (prop) {
        return objectGet(obj, prop) !== undefined
      })
    })
    .map(function (obj) {
      var output = {}
      properties.forEach(function (prop) {
        var lastProp = last(prop.split('.'))
        var value = objectGet(obj, prop)
        if (value) output[lastProp] = value
      })
      return output
    })
}

/**
Works in exactly the same way as {@link module:array-tools.where} but returns only the first item found.

@param {object[]} - the array to search
@param {object} - the search query
@returns {*}
@category not chainable
@static
@example
> dudes = [
    { name: 'Jim', age: 8 },
    { name: 'Clive', age: 8 },
    { name: 'Hater', age: 9 }
]

> a.findWhere(dudes, { age: 8 })
{ name: 'Jim', age: 8 }
*/
function findWhere (recordset, query) {
  return where(recordset, query)[0]
}

/**
Returns an array containing the unique values from the input array.
@param {Array} - input array
@returns {Array}
@category chainable
@example
> a.unique([ 1, 6, 6, 7, 1])
[ 1, 6, 7 ]
@static
*/
function unique (array) {
  return array.reduce(function (prev, curr) {
    if (prev.indexOf(curr) === -1) prev.push(curr)
    return prev
  }, [])
}

/**
Splice items from the input array until the matching test fails. Returns an array containing the items removed.

@param {Array} - the input array
@param {number} - the position to begin splicing from
@param {any} - the sequence of items passing this test will be removed
@param ...elementN {*} - elements to add to the array in place
@returns {Array}
@category chainable
@static
@example
> function under10(n){ return n < 10; }
> numbers = [ 1, 2, 4, 6, 12 ]

> a.spliceWhile(numbers, 0, under10)
[ 1, 2, 4, 6 ]
> numbers
[ 12 ]

> countries = [ "Egypt", "Ethiopia", "France", "Argentina" ]

> a.spliceWhile(countries, 0, /^e/i)
[ 'Egypt', 'Ethiopia' ]
> countries
[ 'France', 'Argentina' ]
*/
function spliceWhile (array, index, test) {
  for (var i = 0; i < array.length; i++) {
    if (!testValue(array[i], test)) break
  }
  var spliceArgs = [ index, i ]
  spliceArgs = spliceArgs.concat(arrayify(arguments).slice(3))
  return array.splice.apply(array, spliceArgs)
}

/**
Removes items from `array` which satisfy the query. Modifies the input array, returns the extracted.

@param {Array} - the input array, modified directly
@param {any} - if an item in the input array passes this test it is removed
@returns {Array} the extracted items.
@category chainable
@static
@example
> DJs = [
    { name: "Trevor", sacked: true },
    { name: "Mike", sacked: true },
    { name: "Chris", sacked: false },
    { name: "Alan", sacked: false }
]

> a.extract(DJs, { sacked: true })
[ { name: 'Trevor', sacked: true },
  { name: 'Mike', sacked: true } ]

> DJs
[ { name: 'Chris', sacked: false },
  { name: 'Alan', sacked: false } ]

*/
function extract (array, query) {
  var result = []
  var toSplice = []
  arrayify(array).forEach(function (item, index) {
    if (testValue(item, query)) {
      result.push(item)
      toSplice.push(index)
    }
  })
  for (var i = 0; i < toSplice.length; i++) {
    array.splice(toSplice[i] - i, 1)
  }
  return result
}

/**
Removes the specified value from the input array.

@param {Array} - the input array
@param {*} - the item to remove
@category not chainable
@return {*}
@static
@since 1.8.0
@example
> numbers = [ 1, 2, 3 ]
> a.remove(numbers, 1)
[ 1 ]

> numbers
[ 2, 3 ]
*/
function remove (arr, toRemove) {
  return arr.splice(arr.indexOf(toRemove), 1)[0]
}

/**
flatten an array of arrays into a single array.

@static
@since 1.4.0
@param {Array} - the input array
@returns {Array}
@category chainable
@example
> numbers = [ 1, 2, [ 3, 4 ], 5 ]
> a.flatten(numbers)
[ 1, 2, 3, 4, 5 ]
*/
function flatten (array) {
  return arrayify(array).reduce(function (prev, curr) {
    return prev.concat(curr)
  }, [])
}

/**
Sort an array of objects by one or more fields
@member sortBy
@static
@param {object[]} - input array
@param {string|string[]} - column name(s) to sort by
@param {object} - specific sort orders, per columns
@returns {Array}
@category chainable
@since 1.5.0
@example
with this data
```js
> DJs = [
    { name: "Trevor", slot: "twilight" },
    { name: "Chris", slot: "twilight" },
    { name: "Mike", slot: "afternoon" },
    { name: "Rodney", slot: "morning" },
    { name: "Chris", slot: "morning" },
    { name: "Zane", slot: "evening" }
]
```

sort by `slot` using the default sort order
```js
> a.sortBy(DJs, "slot")
[ { name: 'Mike', slot: 'afternoon' },
  { name: 'Zane', slot: 'evening' },
  { name: 'Chris', slot: 'morning' },
  { name: 'Rodney', slot: 'morning' },
  { name: 'Chris', slot: 'twilight' },
  { name: 'Trevor', slot: 'twilight' } ]
```

specify a custom sort order for `slot`
```js
> a.sortBy(DJs, "slot", { slot: [ "morning", "afternoon", "evening", "twilight" ]})
[ { name: 'Rodney', slot: 'morning' },
  { name: 'Chris', slot: 'morning' },
  { name: 'Mike', slot: 'afternoon' },
  { name: 'Zane', slot: 'evening' },
  { name: 'Trevor', slot: 'twilight' },
  { name: 'Chris', slot: 'twilight' } ]
```

sort by `slot` then `name`
```js
> a.sortBy(DJs, ["slot", "name"], { slot: [ "morning", "afternoon", "evening", "twilight" ]})
[ { name: 'Chris', slot: 'morning' },
  { name: 'Rodney', slot: 'morning' },
  { name: 'Mike', slot: 'afternoon' },
  { name: 'Zane', slot: 'evening' },
  { name: 'Chris', slot: 'twilight' },
  { name: 'Trevor', slot: 'twilight' } ]
```
*/

/**
Return the last item in an array.
@param {Array} - the input array
@category not chainable
@return {*}
@static
@since 1.7.0
*/
function last (arr) {
  return arr[arr.length - 1]
}

/**
Searches the array for the exact value supplied (strict equality). To query for value existance using an expression or function, use {@link module:array-tools.exists}. If you pass an array of values, contains will return true if they _all_ exist. (note: `exists` returns true if _some_ of them exist).

@param {Array} - the input array
@param {*} - the value to look for
@category not chainable
@return {boolean}
@static
@since 1.8.0
*/
function contains (array, value) {
  if (Array.isArray(array) && !Array.isArray(value)) {
    return array.indexOf(value) > -1
  } else if (Array.isArray(array) && Array.isArray(value)) {
    return value.every(function (item) {
      return contains(array, item)
    })
  } else {
    return array === value
  }
}
