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

var astValue = require('./ast-value');
var analyzeProperties = require('./analyze-properties');

/**
 * Returns an object containing functions that will annotate `declaration` with
 * the polymer-specificmeaning of the value nodes for the named properties.
 *
 * @param  {ElementDescriptor} declaration The descriptor to annotate.
 * @return {object.<string,function>}      An object containing property
 *                                         handlers.
 */
function declarationPropertyHandlers(declaration) {
  return {
    is: function(node) {
      if (node.type == 'Literal') {
        declaration.is = node.value;
      }
    },
    properties: function(node) {

      var props = analyzeProperties(node);

      for (var i=0; i<props.length; i++) {
        declaration.properties.push(props[i]);
      }
    },
    behaviors: function(node) {
      if (node.type != 'ArrayExpression') {
        return;
      }

      for (var i=0; i<node.elements.length; i++) {
        var v = astValue.expressionToValue(node.elements[i]);
        if (v === undefined)
          v = astValue.CANT_CONVERT;
        declaration.behaviors.push(v);
      }
    },
    observers: function(node) {
      if (node.type != 'ArrayExpression') {
        return;
      }
      for (var i=0; i<node.elements.length; i++) {
        var v = astValue.expressionToValue(node.elements[i]);
        if (v === undefined)
          v = astValue.CANT_CONVERT;
        declaration.observers.push({
          javascriptNode: node.elements[i],
          expression: v
        });
      }
    }
  };
}

module.exports = declarationPropertyHandlers;
