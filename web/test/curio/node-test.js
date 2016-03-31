'use strict';

import {assert} from 'chai';
import Node from '../../scripts/curio/node';

describe('Node', function() {
  it('prints a function node', function() {
    var node = Node.function();
    assert.equal(node.print(), '[function argc:0]');

    node = Node.function(null, { argc: 3, infix: true });
    assert.equal(node.print(), '[function argc:3 infix]');

    node = Node.operator();
    assert.equal(node.print(), '[function argc:2 infix]');

    node = Node.operator(function(a) {});
    assert.equal(node.print(), '[function argc:2 infix]');
  });
});
