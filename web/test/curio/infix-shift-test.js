import {assert} from 'chai';
import Node from '../../scripts/curio/node';
import Scope from '../../scripts/curio/scope';
import InfixShift from '../../scripts/curio/infix-shift';

describe('InfixShift', function() {
  it('shifts infix functions left', function() {
    // 1 is 1
    var node = Node.list()
      .integer(1)
      .identifier('is')
      .integer(1)
    .end();

    var scope = new Scope();
    scope.put('is', Node.operator(), 20);

    var infs = new InfixShift(scope);
    infs.transform(node, 20);

    assert.equal(node.print(), 'is 1 1');

    assert.deepEqual(node, Node.list()
      .identifier('is')
      .integer(1)
      .integer(1)
    .end());
  });

  it('shifts infix functions left for a precedence level', function() {
    // 1 is 1 and 2 is 2
    var node = Node.list()
      .integer(1)
      .identifier('is')
      .integer(1)
      .identifier('and')
      .integer(2)
      .identifier('is')
      .integer(2)
    .end();

    var scope = new Scope();
    scope.put('is', Node.operator(), 20);
    scope.put('and', Node.operator(), 30);

    var infs = new InfixShift(scope);
    infs.transform(node, 20);

    assert.equal(node.print(), 'is 1 1 and is 2 2');

    assert.deepEqual(node, Node.list()
      .identifier('is')
      .integer(1)
      .integer(1)
      .identifier('and')
      .identifier('is')
      .integer(2)
      .integer(2)
    .end());
  });
});
