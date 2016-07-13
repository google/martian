'use strict';

export default class ArgumentBinding {
  constructor(scope) {
    this._scope = scope;
  }

  transform(node, precedence) {
    if (!node) {
      return;
    }

    switch (node.type) {
      case 'LIST':
        this.transform(node.down, precedence);
        break;
      case 'IDENTIFIER':
        var scoped = this._scope.get(node.value, precedence);
        if (!scoped) {
          break;
        }

        var sent = node.prev.list();
        var list = sent.end(); 
        var prev = node.prev;

        /*
         * [PREV]<->[NODE]<->[NEXT]
         * [PREV]<->[LIST]<->[NODE]<->[NEXT]
         *            ^-[SENT]
         * [PREV]<->[LIST]<->[NEXT]
         *            ^-[SENT]<->[NODE]
         */
        sent.next = node;
        node.prev = sent;
        list.prev = prev;
        prev.next = list;
        node.up = list;

        var last = node;

        for (var i = 0; i < scoped.argc; i++) {
          last = last.next;
          last.up = list;
        }

        if (last.next) {
          last.next.prev = list;
          list.next = last.next;
          delete(last.next);
        }

        node = list;
    }

    this.transform(node.next, precedence);
  }
}
