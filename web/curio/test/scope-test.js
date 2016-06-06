'use strict';

var chai = require('chai'),
    assert = chai.assert;

var scope = require('../src/scope'),
    parse = require('../src/parser'),
	Scope = scope.Scope,
	Node = parse.Node;

describe('Scope', function() {
    it('maps symbols to nodes', function() {
      var scope = new Scope();
      scope.put('name', Node.string('Curio'));

	  assert.deepEqual(scope.get('name'), Node.string('Curio'));
    });
});
