/**
 * @license
 * Copyright (c) 2015 The Polymer Project Authors. All rights reserved.
 * This code may only be used under the BSD style license found at http://polymer.github.io/LICENSE.txt
 * The complete set of authors may be found at http://polymer.github.io/AUTHORS.txt
 * The complete set of contributors may be found at http://polymer.github.io/CONTRIBUTORS.txt
 * Code distributed by Google as part of the polymer project is also
 * subject to an additional IP rights grant found at http://polymer.github.io/PATENTS.txt
 */
/**
* Finds and annotates the Polymer() and modulate() calls in javascript.
*/
// jshint node: true
'use strict';
var espree = require('espree');
var estraverse = require('estraverse');

var behaviorFinder = require('./behavior-finder');
var elementFinder  = require('./element-finder');
var featureFinder  = require('./feature-finder');

function traverse(visitorRegistries) {
  var visitor;
  function applyVisitors(name, node, parent) {
    var returnVal;
    for (var i = 0; i < visitorRegistries.length; i++) {
      if (name in visitorRegistries[i]) {
        returnVal = visitorRegistries[i][name](node, parent);
        if (returnVal) {
          return returnVal;
        }
      }
    }
  }
  return {
    enter: function(node, parent) {
      visitor = 'enter' + node.type;
      return applyVisitors(visitor, node, parent);
    },
    leave: function(node, parent) {
      visitor = 'leave' + node.type;
      return applyVisitors(visitor, node, parent);
    },
    fallback: 'iteration',
  };
}

var jsParse = function jsParse(jsString) {
  var script = espree.parse(jsString, {
    attachComment: true,
    comment: true,
    loc: true,
    ecmaFeatures: {
      arrowFunctions: true,
      blockBindings: true,
      destructuring: true,
      regexYFlag: true,
      regexUFlag: true,
      templateStrings: true,
      binaryLiterals: true,
      unicodeCodePointEscapes: true,
      defaultParams: true,
      restParams: true,
      forOf: true,
      objectLiteralComputedProperties: true,
      objectLiteralShorthandMethods: true,
      objectLiteralShorthandProperties: true,
      objectLiteralDuplicateProperties: true,
      generators: true,
      spread: true,
      classes: true,
      modules: true,
      jsx: true,
      globalReturn: true,
    }
  });

  var featureInfo = featureFinder();
  var behaviorInfo = behaviorFinder();
  var elementInfo = elementFinder();

  var visitors = [featureInfo, behaviorInfo, elementInfo].map(function(info) {
    return info.visitors;
  });
  estraverse.traverse(script, traverse(visitors));

  return {
    behaviors: behaviorInfo.behaviors,
    elements:  elementInfo.elements,
    features:  featureInfo.features,
    parsedScript: script
  };
};

module.exports = jsParse;
