"use strict";
var arrayify = require("array-back");

/**
Sort an array of objects by any number of fields, in any order. 

@module sort-array
@typicalname sortBy
@example
var sortBy = require("sort-array");
*/
module.exports = sortBy;

/**
Sort an array of objects by one or more fields
@param {object[]} - input array
@param {string|string[]} - column name(s) to sort by
@param {object} - specific sort orders, per columns
@returns {Array}
@alias module:sort-array
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
*/
function sortBy(recordset, columns, customOrder){
    return recordset.sort(sortByFunc(arrayify(columns), customOrder));
}

function sortByFunc(properties, customOrder){
    var props = properties.slice(0);
    var property = props.shift();
    return function tryIt(a, b){
        var result;
        var x = a[property];
        var y = b[property];

        if (typeof x === "undefined" && typeof y !== "undefined"){
            result = -1;
        } else if (typeof x !== "undefined" && typeof y === "undefined"){
            result = 1;
        } else if (typeof x === "undefined" && typeof y === "undefined"){
            result = 0;
        } else if (customOrder && customOrder[property]){
            result = customOrder[property].indexOf(x) - customOrder[property].indexOf(y);
        } else {
            result = x < y ? -1 : x > y ? 1 : 0;
        }

        if (result === 0){
            if (props.length){
                property = props.shift();
                return tryIt(a, b);
            } else {
                return 0;
            }
        } else {
            props = properties.slice(0);
            property = props.shift();
            return result;
        }
        return 0;
    };
}