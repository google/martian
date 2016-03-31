/**
 * @license
 * Copyright (c) 2015 The Polymer Project Authors. All rights reserved.
 * This code may only be used under the BSD style license found at http://polymer.github.io/LICENSE.txt
 * The complete set of authors may be found at http://polymer.github.io/AUTHORS.txt
 * The complete set of contributors may be found at http://polymer.github.io/CONTRIBUTORS.txt
 * Code distributed by Google as part of the polymer project is also
 * subject to an additional IP rights grant found at http://polymer.github.io/PATENTS.txt
 */
// jshint node: true
'use strict';
var estraverse = require("estraverse");

/**
 * Returns whether an Espree node matches a particular object path.
 *
 * e.g. you have a MemberExpression node, and want to see whether it represents
 * `Foo.Bar.Baz`:
 *
 *     matchesCallExpression(node, ['Foo', 'Bar', 'Baz'])
 *
 * @param {Node} expression The Espree node to match against.
 * @param {Array<string>} path The path to look for.
 */
function matchesCallExpression(expression, path) {
  if (!expression.property || !expression.object) return;
  console.assert(path.length >= 2);

  // Unravel backwards, make sure properties match each step of the way.
  if (expression.property.name !== path[path.length - 1]) return false;
  // We've got ourselves a final member expression.
  if (path.length == 2 && expression.object.type === 'Identifier') {
    return expression.object.name === path[0];
  }
  // Nested expressions.
  if (path.length > 2 && expression.object.type == 'MemberExpression') {
    return matchesCallExpression(expression.object, path.slice(0, path.length - 1));
  }

  return false;
}

/**
 * @param {Node} key The node representing an object key or expression.
 * @return {string} The name of that key.
 */
function objectKeyToString(key) {
  if (key.type == 'Identifier') {
    return key.name;
  }
  if (key.type == 'Literal') {
    return key.value;
  }
  if (key.type == 'MemberExpression') {
    return objectKeyToString(key.object) + '.' + objectKeyToString(key.property);
  }
}

var CLOSURE_CONSTRUCTOR_MAP = {
  'Boolean': 'boolean',
  'Number':  'number',
  'String':  'string',
};

/**
 * AST expression -> Closure type.
 *
 * Accepts literal values, and native constructors.
 *
 * @param {Node} node An Espree expression node.
 * @return {string} The type of that expression, in Closure terms.
 */
function closureType(node) {
  if (node.type.match(/Expression$/)) {
    return node.type.substr(0, node.type.length - 10);
  } else if (node.type === 'Literal') {
    return typeof node.value;
  } else if (node.type === 'Identifier') {
    return CLOSURE_CONSTRUCTOR_MAP[node.name] || node.name;
  } else {
    throw {
      message: 'Unknown Closure type for node: ' + node.type,
      location: node.loc.start,
    };
  }
}

/**
 * @param {Node} node
 * @return {?string}
 */
function getAttachedComment(node) {
  var comments = getLeadingComments(node) || getLeadingComments(node.key);
  if (!comments) {
    return;
  }
  return comments[comments.length - 1];
}

/**
 * Returns all comments from a tree defined with @event.
 * @param  {Node} node [description]
 * @return {[type]}      [description]
 */
function getEventComments(node) {
  var eventComments = [];
  estraverse.traverse(node, {
    enter: function (node) {
      var comments = (node.leadingComments || []).concat(node.trailingComments || [])
        .map( function(commentAST) {
          return commentAST.value;
        })
        .filter( function(comment) {
          return comment.indexOf("@event") != -1;
        });
      eventComments = eventComments.concat(comments);
    }
  });
  // dedup
  return eventComments.filter( function(el, index, array) {
    return array.indexOf(el) === index;
  });
}

/**
 * @param {Node} node
 * @param
 * @return {Array.<string>}
 */
function getLeadingComments(node) {
  if (!node) {
    return;
  }
  var comments = node.leadingComments;
  if (!comments || comments.length === 0) return;
  return comments.map(function(comment) {
    return comment.value;
  });
}

/**
 * Converts a parse5 Property AST node into its Hydrolysis representation.
 *
 * @param {Node} node
 * @return {PropertyDescriptor}
 */
function toPropertyDescriptor(node) {
  var type = closureType(node.value);
  if (type == "Function") {
    if (node.kind === "get" || node.kind === "set") {
      type = '';
      node[node.kind+"ter"] = true;
    }
  }
  var result = {
    name: objectKeyToString(node.key),
    type: type,
    desc: getAttachedComment(node),
    javascriptNode: node
  };

  if (type === 'Function') {
    result.params = (node.value.params || []).map(function(param) {
      return {name: param.name};
    });
  }

  return result;
}

module.exports = {
  closureType:           closureType,
  getAttachedComment:    getAttachedComment,
  getEventComments:      getEventComments,
  matchesCallExpression: matchesCallExpression,
  objectKeyToString:     objectKeyToString,
  toPropertyDescriptor:  toPropertyDescriptor,
};
