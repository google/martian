'use strict';

import {assert} from 'chai';
import Curio from '../../scripts/curio/curio';

describe('Curio', function() {
  it('runs a simple filter', function() {
    var c = new Curio();

    c.compile('1 is 1 and 2 is 2 and 3 is 3');

    console.log(c._node.print());

    assert.equal(c.run({}), true);
  });
});
