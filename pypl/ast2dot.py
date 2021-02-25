# -*-coding: utf8 -*-

"""
Dump an ast tree in dot languae

"""

import ast

def dump(node):
    def _format(node):
        if not isinstance(node, ast.AST):
            return ['%d [label="%s\\n%s"];' % (id(node),
                node.__class__.__name__, str(node))]
        c = []
        fields = [b for a, b in ast.iter_fields(node)]
        for n, f in ast.iter_fields(node):
            if isinstance(f, list):
                for b in f:
                    c.append('%d -> %d [label=%s];' % (id(node), id(b), n))
                    c.extend(_format(b))
            else:
                c.append('%d -> %d [label=%s];' % (id(node), id(f), n))
                c.extend(_format(f))

        attrs = filter(lambda x: x != 'lineno' and x!= 'col_offset', 
                       node._attributes)
        label = '\\n'.join('%s=%s' % (a, str(getattr(node, a))) for a in attrs)
        c.append('%d [label="%s\\n%s"];' % (id(node),
                    node.__class__.__name__, label))
        return c

    if not isinstance(node, ast.AST):
        raise TypeError('expected AST, got %r' % node.__class__.__name__)
    return '\n'.join(['digraph G {\nfontsize=8;', '\n'.join(_format(node)), '}'])

if __name__ == '__main__':
    import sys
    filename = sys.argv[1]
    from subprocess import Popen, PIPE
    p = Popen(['dot', '-Tpng'], stdin=PIPE, stdout=PIPE)
    outp, _ = p.communicate(dump(ast.parse(open(filename).read())))
    open(filename + '.png', 'wb').write(outp)

