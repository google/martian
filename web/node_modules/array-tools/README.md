[![view on npm](http://img.shields.io/npm/v/array-tools.svg)](https://www.npmjs.org/package/array-tools)
[![npm module downloads per month](http://img.shields.io/npm/dm/array-tools.svg)](https://www.npmjs.org/package/array-tools)
[![Build Status](https://travis-ci.org/75lb/array-tools.svg?branch=master)](https://travis-ci.org/75lb/array-tools)
[![Dependency Status](https://david-dm.org/75lb/array-tools.svg)](https://david-dm.org/75lb/array-tools)
[![Coverage Status](https://coveralls.io/repos/75lb/array-tools/badge.svg?branch=master)](https://coveralls.io/r/75lb/array-tools?branch=master)
[![js-standard-style](https://img.shields.io/badge/code%20style-standard-brightgreen.svg)](https://github.com/feross/standard)

# array-tools
Lightweight, use-anywhere toolkit for working with array data.

There are four ways to use it.

1) As a command-line tool. E.g. array-tools downloads last month:
```sh
$ curl -s https://api.npmjs.org/downloads/range/last-month/array-tools \
| object-tools get downloads \
| array-tools pluck downloads \
| array-tools join "," \
| spark
▂▅▃▅▅▁▁▃▄▃▆▂▂▁▁▂▄▃▃▁▁▂█▆▆▄▁▃▅▃
```

2) As a standard library, passing the input array on each method invocation:

```js
> var a = require("array-tools");

> var remainder = a.without([ 1, 2, 3, 4, 5 ], 1)
> a.exists(remainder, 1)
false
```

3) As a chainable method, passing the input array once then chaining from there:

```js
> a([ 1, 2, 3, 4, 5 ]).without(1).exists(1);
false
```

4) As a base class.
```js
var util = require("util");
var ArrayTools = require("array-tools");

// this class will inherit all array-tools methods
function CarCollection(cars){
  ArrayTools.call(this, cars);
}
util.inherits(CarCollection, ArrayTools);

var cars = new CarCollection([
  { owner: "Me", model: "Citreon Xsara" },
  { owner: "Floyd", model: "Bugatti Veyron" }
]);

cars.findWhere({ owner: "Floyd" });
// returns { owner: "Floyd", model: "Bugatti Veyron" }
```

#### More on chaining
* Each method returning an `Array` (e.g. `where`, `without`) can be chained.
* Methods not returning an array (`exists`, `contains`) cannot be chained.
* All methods from `Array.prototype` (e.g. `.join`, `.forEach` etc.) are also available in the chain. The same rules, regarding what can and cannot be chained, apply as above.
* If the final operation in your chain is "chainable" (returns an array), append `.val()` to terminate the chain and retrieve the output.

```js
> a([ 1, 2, 2, 3 ]).exists(1)
true
> a([ 1, 2, 2, 3 ]).without(1).exists(1)
false
> a([ 1, 2, 2, 3 ]).without(1).unique().val()
[ 2, 3 ]
> a([ 1, 2, 2, 3 ]).without(1).unique().join("-")
'2-3'
```

## Compatibility
This library is tested in node versions 0.10, 0.11, 0.12, iojs and the following browsers:

[![Sauce Test Status](https://saucelabs.com/browser-matrix/arr-tools.svg)](https://saucelabs.com/u/arr-tools)

## Install
As a library:

```
$ npm install array-tools --save
```

As a command-line tool:
```
$ npm install -g array-tools
```

Using bower:
```
$ bower install array-tools --save
```

## API Reference

* [array-tools](#module_array-tools)
  * _chainable_
    * [.sortBy](#module_array-tools.sortBy) ⇒ <code>Array</code>
    * [.arrayify(any)](#module_array-tools.arrayify) ⇒ <code>Array</code>
    * [.where(array, query)](#module_array-tools.where) ⇒ <code>Array</code>
    * [.without(array, toRemove)](#module_array-tools.without) ⇒ <code>Array</code>
    * [.pluck(recordset, property)](#module_array-tools.pluck) ⇒ <code>Array</code>
    * [.pick(recordset, property)](#module_array-tools.pick) ⇒ <code>Array.&lt;object&gt;</code>
    * [.unique(array)](#module_array-tools.unique) ⇒ <code>Array</code>
    * [.spliceWhile(array, index, test, ...elementN)](#module_array-tools.spliceWhile) ⇒ <code>Array</code>
    * [.extract(array, query)](#module_array-tools.extract) ⇒ <code>Array</code>
    * [.flatten(array)](#module_array-tools.flatten) ⇒ <code>Array</code>
  * _not chainable_
    * [.exists(array, query)](#module_array-tools.exists) ⇒ <code>boolean</code>
    * [.findWhere(recordset, query)](#module_array-tools.findWhere) ⇒ <code>\*</code>
    * [.remove(arr, toRemove)](#module_array-tools.remove) ⇒ <code>\*</code>
    * [.last(arr)](#module_array-tools.last) ⇒ <code>\*</code>
    * [.contains(array, value)](#module_array-tools.contains) ⇒ <code>boolean</code>

<a name="module_array-tools.sortBy"></a>
### a.sortBy ⇒ <code>Array</code>
Sort an array of objects by one or more fields

**Kind**: static property of <code>[array-tools](#module_array-tools)</code>  
**Category**: chainable  
**Since**: 1.5.0  

| Type | Description |
| --- | --- |
| <code>Array.&lt;object&gt;</code> | input array |
| <code>string</code> &#124; <code>Array.&lt;string&gt;</code> | column name(s) to sort by |
| <code>object</code> | specific sort orders, per columns |

**Example**  
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
<a name="module_array-tools.arrayify"></a>
### a.arrayify(any) ⇒ <code>Array</code>
Takes any input and guarantees an array back.

- converts array-like objects (e.g. `arguments`) to a real array
- converts `undefined` to an empty array
- converts any another other, singular value (including `null`) into an array containing that value
- ignores input which is already an array

**Kind**: static method of <code>[array-tools](#module_array-tools)</code>  
**Category**: chainable  

| Param | Type | Description |
| --- | --- | --- |
| any | <code>\*</code> | the input value to convert to an array |

**Example**  
```js
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
```
<a name="module_array-tools.where"></a>
### a.where(array, query) ⇒ <code>Array</code>
Deep query an array.

**Kind**: static method of <code>[array-tools](#module_array-tools)</code>  
**Category**: chainable  

| Param | Type | Description |
| --- | --- | --- |
| array | <code>Array.&lt;object&gt;</code> | the array to query |
| query | <code>any</code> &#124; <code>Array.&lt;any&gt;</code> | one or more queries |

**Example**  
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
<a name="module_array-tools.without"></a>
### a.without(array, toRemove) ⇒ <code>Array</code>
Returns a new array with the same content as the input minus the specified values. It accepts the same query syntax as [where](#module_array-tools.where).

**Kind**: static method of <code>[array-tools](#module_array-tools)</code>  
**Category**: chainable  

| Param | Type | Description |
| --- | --- | --- |
| array | <code>Array</code> | the input array |
| toRemove | <code>any</code> &#124; <code>Array.&lt;any&gt;</code> | one, or more queries |

**Example**  
```js
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
```
<a name="module_array-tools.pluck"></a>
### a.pluck(recordset, property) ⇒ <code>Array</code>
Returns an array containing each value plucked from the specified property of each object in the input array.

**Kind**: static method of <code>[array-tools](#module_array-tools)</code>  
**Category**: chainable  

| Param | Type | Description |
| --- | --- | --- |
| recordset | <code>Array.&lt;object&gt;</code> | The input recordset |
| property | <code>string</code> &#124; <code>Array.&lt;string&gt;</code> | Property name, or an array of property names. If an array is supplied, the first existing property will be returned. |

**Example**  
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
<a name="module_array-tools.pick"></a>
### a.pick(recordset, property) ⇒ <code>Array.&lt;object&gt;</code>
return a copy of the input `recordset` containing objects having only the cherry-picked properties

**Kind**: static method of <code>[array-tools](#module_array-tools)</code>  
**Category**: chainable  

| Param | Type | Description |
| --- | --- | --- |
| recordset | <code>Array.&lt;object&gt;</code> | the input |
| property | <code>string</code> &#124; <code>Array.&lt;string&gt;</code> | the properties to include in the result |

**Example**  
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
<a name="module_array-tools.unique"></a>
### a.unique(array) ⇒ <code>Array</code>
Returns an array containing the unique values from the input array.

**Kind**: static method of <code>[array-tools](#module_array-tools)</code>  
**Category**: chainable  

| Param | Type | Description |
| --- | --- | --- |
| array | <code>Array</code> | input array |

**Example**  
```js
> a.unique([ 1, 6, 6, 7, 1])
[ 1, 6, 7 ]
```
<a name="module_array-tools.spliceWhile"></a>
### a.spliceWhile(array, index, test, ...elementN) ⇒ <code>Array</code>
Splice items from the input array until the matching test fails. Returns an array containing the items removed.

**Kind**: static method of <code>[array-tools](#module_array-tools)</code>  
**Category**: chainable  

| Param | Type | Description |
| --- | --- | --- |
| array | <code>Array</code> | the input array |
| index | <code>number</code> | the position to begin splicing from |
| test | <code>any</code> | the sequence of items passing this test will be removed |
| ...elementN | <code>\*</code> | elements to add to the array in place |

**Example**  
```js
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
```
<a name="module_array-tools.extract"></a>
### a.extract(array, query) ⇒ <code>Array</code>
Removes items from `array` which satisfy the query. Modifies the input array, returns the extracted.

**Kind**: static method of <code>[array-tools](#module_array-tools)</code>  
**Returns**: <code>Array</code> - the extracted items.  
**Category**: chainable  

| Param | Type | Description |
| --- | --- | --- |
| array | <code>Array</code> | the input array, modified directly |
| query | <code>any</code> | if an item in the input array passes this test it is removed |

**Example**  
```js
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
```
<a name="module_array-tools.flatten"></a>
### a.flatten(array) ⇒ <code>Array</code>
flatten an array of arrays into a single array.

**Kind**: static method of <code>[array-tools](#module_array-tools)</code>  
**Category**: chainable  
**Since**: 1.4.0  

| Param | Type | Description |
| --- | --- | --- |
| array | <code>Array</code> | the input array |

**Example**  
```js
> numbers = [ 1, 2, [ 3, 4 ], 5 ]
> a.flatten(numbers)
[ 1, 2, 3, 4, 5 ]
```
<a name="module_array-tools.exists"></a>
### a.exists(array, query) ⇒ <code>boolean</code>
Works in exactly the same way as [where](#module_array-tools.where) but returning a boolean indicating whether a matching record exists.

**Kind**: static method of <code>[array-tools](#module_array-tools)</code>  
**Category**: not chainable  

| Param | Type | Description |
| --- | --- | --- |
| array | <code>Array</code> | the array to search |
| query | <code>\*</code> | the value to search for |

**Example**  
```js
> data = [
    { name: "Dana", age: 30 },
    { name: "Yana", age: 20 },
    { name: "Zhana", age: 10 }
]

> a.exists(data, { age: 10 })
true
```
<a name="module_array-tools.findWhere"></a>
### a.findWhere(recordset, query) ⇒ <code>\*</code>
Works in exactly the same way as [where](#module_array-tools.where) but returns only the first item found.

**Kind**: static method of <code>[array-tools](#module_array-tools)</code>  
**Category**: not chainable  

| Param | Type | Description |
| --- | --- | --- |
| recordset | <code>Array.&lt;object&gt;</code> | the array to search |
| query | <code>object</code> | the search query |

**Example**  
```js
> dudes = [
    { name: 'Jim', age: 8 },
    { name: 'Clive', age: 8 },
    { name: 'Hater', age: 9 }
]

> a.findWhere(dudes, { age: 8 })
{ name: 'Jim', age: 8 }
```
<a name="module_array-tools.remove"></a>
### a.remove(arr, toRemove) ⇒ <code>\*</code>
Removes the specified value from the input array.

**Kind**: static method of <code>[array-tools](#module_array-tools)</code>  
**Category**: not chainable  
**Since**: 1.8.0  

| Param | Type | Description |
| --- | --- | --- |
| arr | <code>Array</code> | the input array |
| toRemove | <code>\*</code> | the item to remove |

**Example**  
```js
> numbers = [ 1, 2, 3 ]
> a.remove(numbers, 1)
[ 1 ]

> numbers
[ 2, 3 ]
```
<a name="module_array-tools.last"></a>
### a.last(arr) ⇒ <code>\*</code>
Return the last item in an array.

**Kind**: static method of <code>[array-tools](#module_array-tools)</code>  
**Category**: not chainable  
**Since**: 1.7.0  

| Param | Type | Description |
| --- | --- | --- |
| arr | <code>Array</code> | the input array |

<a name="module_array-tools.contains"></a>
### a.contains(array, value) ⇒ <code>boolean</code>
Searches the array for the exact value supplied (strict equality). To query for value existance using an expression or function, use [exists](#module_array-tools.exists). If you pass an array of values, contains will return true if they _all_ exist. (note: `exists` returns true if _some_ of them exist).

**Kind**: static method of <code>[array-tools](#module_array-tools)</code>  
**Category**: not chainable  
**Since**: 1.8.0  

| Param | Type | Description |
| --- | --- | --- |
| array | <code>Array</code> | the input array |
| value | <code>\*</code> | the value to look for |


* * *

&copy; 2015 Lloyd Brookes <75pound@gmail.com>. Documented by [jsdoc-to-markdown](https://github.com/75lb/jsdoc-to-markdown).
