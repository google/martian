'use strict';

var chai = require('chai'),
    assert = chai.assert;

var evaluate = require('../../scripts/curio/evaluator'),
    parse = require('../../scripts/curio/parser'),
    Evaluator = evaluate.Evaluator,
    Node = parse.Node;

import Scope from '../../scripts/curio/scope';

describe('Evaluator', function() {
  it('evaluates a simple expression', function() {
    var scope = new Scope();

    scope.put('is', Node.function(2, true, function(a, b) {
      return Node.boolean(a.value == b.value);
    }));

    var evaluator = new Evaluator({}, scope);

    var root = Node.list()
      .identifier('is')
      .integer(1)
      .integer(1)
    .end();

    var result = evaluator.evaluate(root);

    assert.equal(result.print(), 'true');

    assert.deepEqual(result, Node.list().boolean(true).end());

    root = Node.list()
      .identifier('is')
      .integer(1)
      .integer(2)
    .end();

    result = evaluator.evaluate(root);

    assert.equal(result.print(), 'false');

    assert.deepEqual(result, Node.list().boolean(false).end());
  });
});
