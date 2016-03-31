[![view on npm](http://img.shields.io/npm/v/sort-array.svg)](https://www.npmjs.org/package/sort-array)
[![npm module downloads per month](http://img.shields.io/npm/dm/sort-array.svg)](https://www.npmjs.org/package/sort-array)
[![Build Status](https://travis-ci.org/75lb/sort-array.svg?branch=master)](https://travis-ci.org/75lb/sort-array)
[![Dependency Status](https://david-dm.org/75lb/sort-array.svg)](https://david-dm.org/75lb/sort-array)

<a name="module_sort-array"></a>
## sort-array
Sort an array of objects by any number of fields, in any order.

**Example**  
```js
var sortBy = require("sort-array");
```
<a name="exp_module_sort-array--sortBy"></a>
### sortBy(recordset, columns, customOrder) ⇒ <code>Array</code> ⏏
Sort an array of objects by one or more fields

**Kind**: Exported function  

| Param | Type | Description |
| --- | --- | --- |
| recordset | <code>Array.&lt;object&gt;</code> | input array |
| columns | <code>string</code> &#124; <code>Array.&lt;string&gt;</code> | column name(s) to sort by |
| customOrder | <code>object</code> | specific sort orders, per columns |

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

sort by `slot` using the default sort order (alphabetical)
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
> var slotOrder = [ "morning", "afternoon", "evening", "twilight" ];
> a.sortBy(DJs, "slot", { slot: slotOrder })
[ { name: 'Rodney', slot: 'morning' },
  { name: 'Chris', slot: 'morning' },
  { name: 'Mike', slot: 'afternoon' },
  { name: 'Zane', slot: 'evening' },
  { name: 'Trevor', slot: 'twilight' },
  { name: 'Chris', slot: 'twilight' } ]
```

sort by `slot` then `name`
```js
> a.sortBy(DJs, ["slot", "name"], { slot: slotOrder })
[ { name: 'Chris', slot: 'morning' },
  { name: 'Rodney', slot: 'morning' },
  { name: 'Mike', slot: 'afternoon' },
  { name: 'Zane', slot: 'evening' },
  { name: 'Chris', slot: 'twilight' },
  { name: 'Trevor', slot: 'twilight' } ]
```

* * *

&copy; 2015 Lloyd Brookes \<75pound@gmail.com\>. Documented by [jsdoc-to-markdown](https://github.com/jsdoc2md/jsdoc-to-markdown).
