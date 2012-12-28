# -*-coding: utf8 -*-

"""
Random string generator based on regex pattern

(?iLmsux) (?=...) (?!...) (?<=...) (?<!...) are not supported, 'cause i'm not
sure if the lookahead make any sense in the generation process.

All the anchors in the pattern is ignored if they present.

ref: http://docs.python.org/library/re.html#regular-expression-syntax

>>> import re
>>> def _(p):
...     return re.match(p, gen(p)) is not None
>>> _('[a-z]{1,4}a+b?c*')
True
>>> _('([1-9]|1[0-2]|0[1-9]){1}(:[0-5][0-9][ap][m]){1}')
True
>>> _('(?P<mark>a|b|c)(?P=mark)')
True
>>> _('((?:(?:25[0-5]|2[0-4]\d|[01]?\d?\d)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d?\d))')
True
"""

import random
import string
import sre_parse
from sre_constants import (
    MAXREPEAT, LITERAL, NEGATE, CATEGORY,
)

# the last 5 is '\t\n\r\x0b\x0c'
ANY_CHAR = string.printable[:-5]
DIGIT = string.digits
SPACE = ' \t'
WORD = string.letters + string.digits + '_'
NOT_DIGIT = string.letters + string.punctuation + SPACE
NOT_SPACE = string.printable[:-6]
NOT_WORD = """!"#$%&'()*+,-./:;<=>?@[\]^`{|}~"""

def _set(item):
    return globals().get(item.split('_', 1)[1].upper(), '')

class Generator:
    def __init__(self, str):
        self.p = sre_parse.parse(str)

    def gen(self):
        self._groups = {}
        return self._(self.p)

    def _(self, p):
        e = []
        for node in p:
            try:
                e.append(getattr(self, '_'+node[0])(node[1]))
            except AttributeError:
                pass
        return ''.join(e)

    def _literal(self, item):
        return chr(item)

    def _not_literal(self, item):
        s = random.sample(ANY_CHAR, 2)
        return s[0] if s[1] == item else s[1]

    def _max_repeat(self, item):
        stop = item[1] if item[1] != MAXREPEAT else item[0] + 20
        n = random.randint(item[0], stop)
        return ''.join([self._(item[2]) for i in range(n)])

    def _min_repeat(self, item):
        # non-greedy repeatition
        return ''.join([self._(item[2]) for i in range(item[0])])

    def _in(self, item):
        if item[0][0] == NEGATE:
            # negative set
            ns = set()
            for i in item[1:]:
                if i[0] == LITERAL:
                    ns.add(chr(i[1]))
                elif i[0] == CATEGORY:
                    ns.update(_set(i[1]))
                else:
                    start, stop = i[1]
                    ns.update([chr(d) for d in range(start, stop+1)])
            return random.choice(list(set(ANY_CHAR) - ns))
        else:
            # positive set
            return self._([random.choice(item)])

    def _range(self, item):
        return chr(random.randint(*item))

    def _any(self, item):
        return random.choice(ANY_CHAR)

    def _category(self, item):
        s = _set(item)
        return random.choice(s) if s else ''

    def _subpattern(self, item):
        s = self._(item[1])
        # save for backreference
        if item[0] is not None:
            self._groups[item[0]] = s
        return s

    def _branch(self, item):
        return self._(random.choice(item[1]))

    def _groupref(self, item):
        return self._groups.get(item, '')

    def _groupref_exists(self, item):
        return self._(item[1] if self._groups.has_key(item[0]) else item[2])

def gen(str):
    return Generator(str).gen()

if __name__ == '__main__':
    import sys; print gen(sys.argv[1])
