'use strict';

import {assert} from 'chai';
import Node from '../../scripts/curio/node';
import Scope from '../../scripts/curio/scope';
import ArgumentBinding from '../../scripts/curio/argument-binding';

describe('ArgumentBinding', function() {
  it('binds arguments to a function call', function() {
    // "definitely" is 1 1 always
    var node = Node.list()
      .string('definitely')
      .identifier('is')
      .integer(1)
      .integer(1)
      .identifier('always')
    .end();

    var scope = new Scope();
    scope.put('is', Node.operator(), 20);

    var argb = new ArgumentBinding(scope);

    argb.transform(node, 20);

    assert.equal(node.print(), '"definitely" (is 1 1) always');

    assert.deepEqual(node, Node.list()
      .string('definitely')
      .list()
        .identifier('is')
        .integer(1)
        .integer(1)
      .end()
      .identifier('always')
    .end());
  });
});
