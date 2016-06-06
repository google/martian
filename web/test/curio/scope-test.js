'use strict';

var chai = require('chai'),
    assert = chai.assert;

import Scope from '../../scripts/curio/scope';
var parse = require('../../scripts/curio/parser'),
    Node = parse.Node;

describe('Scope', function() {
  it('maps symbols to nodes', function() {
    var scope = new Scope();

    scope.put('name', Node.string('Curio'));

    assert.deepEqual(scope.get('name'), Node.string('Curio'));
  });

  it('maintains a stack of scopes', function() {
    var scope = new Scope();
    scope.put('name', Node.string('Curio'));

    scope = scope.push();
    scope.put('test', Node.boolean(true));

    assert.deepEqual(scope.get('name'), Node.string('Curio'));
    assert.deepEqual(scope.get('test'), Node.boolean(true));

    scope = scope.pop();

    assert.isNull(scope.get('test'));
    assert.deepEqual(scope.get('name'), Node.string('Curio'));
  });
});
