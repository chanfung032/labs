# -*-coding: utf8 -*- 

"""A compiler for PL/0 language
"""

from pyparsing import *

LPAR, RPAR, SEMI, COMMA, PERIOD = map(Suppress, '();,.')

keyword = ('program', 'function', 'var', 'const', 'begin', 'end', 'return',
           'while', 'do', 'if', 'then', 'else', 'odd', 'read', 'write')
__dict__ = locals()
for k in keyword:
    __dict__[k.upper()] = Keyword(k)

ident     = Word(alphas+'_', alphanums+'_')
integer   = Word(nums, nums)
string_   = quotedString

aop       = Literal('+') ^ Literal('-')
mop       = Literal('*') ^ Literal('/')
lop       = Literal('=') ^ Literal('<>') ^ Literal('<') ^ Literal('<=') ^ \
            Literal('>') ^ Literal('>=')
uop       = Literal('+') ^ Literal('-')

expr      = Forward()
call      = ident + LPAR + Optional(delimitedList(expr)) + RPAR
factor    = call | (ident ^ integer ^ string_ ^ LPAR + expr + RPAR)
term      = factor + ZeroOrMore(mop + factor)
expr      << Optional(uop) + term + ZeroOrMore(aop + term)
cond      = expr + lop + expr ^ ODD + expr

body      = Forward()
stmt      = Forward()
stmt      <<(ident + ':=' + expr ^ \
             IF + cond + THEN + stmt + Optional(ELSE + stmt) ^ \
             WHILE + cond + DO + stmt ^ \
             RETURN + expr ^ \
             body ^ \
             (WRITE + LPAR + delimitedList(expr) + RPAR | expr))
body      << BEGIN + stmt + ZeroOrMore(SEMI + stmt) + END

vardecl   = VAR + ident + ZeroOrMore(COMMA + ident) + SEMI
const     = ident + Suppress('=') + (integer ^ string_)
constdecl = CONST + const + ZeroOrMore(COMMA + const) + SEMI

func      = Forward()
block     = Optional(constdecl) + Optional(vardecl) + Optional(func) + body

func      << FUNCTION + ident + \
             LPAR + Group(Optional(delimitedList(ident))) + RPAR + \
             SEMI + block + ZeroOrMore(SEMI + func)

program   = PROGRAM + ident + SEMI + block + PERIOD

comment   = Regex(r"\{[^}]*?\}")
program.ignore(comment)

import ast

def _i(loc, s):
    return {'lineno': lineno(loc, s), 'col_offset': col(loc, s)}

def _list(t):
    return t if isinstance(t, list) else [t]

def do_ident(s, loc, toks):
    ident = ast.Name(**_i(loc, s))
    ident.id = toks[0]
    # ident.ctx = None, here we do not know the context
    return [ident]

def do_integer(s, loc, toks):
    return [ast.Num(int(toks[0]), **_i(loc, s))]

def do_string_(s, loc, toks):
    return [ast.Str(toks[0][1:-1], **_i(loc, s))]

def do_aop(s, loc, toks):
    return [ast.Add(**_i(loc, s)) if toks[0] == '+' else ast.Sub(**_i(loc, s))]

def do_uop(s, loc, toks):
    return [ast.UAdd(**_i(loc, s)) if toks[0] == '+' else ast.USub(**_i(loc, s))]

def do_mop(s, loc, toks):
    return [ast.Mult(**_i(loc, s)) if toks[0] == '*' else ast.Div(**_i(loc, s))]

def do_lop(s, loc, toks):
    _mapping = { 
        '=': ast.Eq, '<>': ast.NotEq, '<': ast.Lt, '<=': ast.LtE,
        '>': ast.Gt, '>=': ast.GtE
    }
    return [_mapping[toks[0]](**_i(loc, s))]

def do_const(s, loc, toks):
    toks[0].ctx = ast.Store()
    return [ast.Assign([toks[0],], toks[1], **_i(loc, s))]

def do_constdecl(s, loc, toks):
    return toks[1:] # ignore the leading 'const' keyword

def do_call(s, loc, toks):
    toks[0].ctx = ast.Load()
    return [ast.Call(toks[0], toks[1:], [], None, None, **_i(loc, s))]

def do_factor(s, loc, toks):
    for t in toks:
        if isinstance(t, ast.Name):
            # FIXME: validate this symbol
            t.ctx = ast.Load()

def do_term(s, loc, toks):
    left = toks[0]
    for i in range(1, len(toks), 2):
        left = ast.BinOp(left, toks[i], toks[i+1], **_i(loc, s))
    return [left]

def do_expr(s, loc, toks):
    if isinstance(toks[0], ast.USub) or isinstance(toks[0], ast.UAdd):
        left = ast.UnaryOp(toks[0], toks[1], **_i(loc, s))
        start = 2
    else:
        left = toks[0]
        start = 1
    for i in range(start, len(toks), 2):
        left = ast.BinOp(left, toks[i], toks[i+1], **_i(loc, s))
    return [left]

def do_cond(s, loc, toks):
    if len(toks) == 2:
        # odd integer: integer % 2 != 0
        return ast.Compare(ast.BinOp(toks[1], ast.Mod(), ast.Num(2)), 
                           [ast.NotEq()], [ast.Num(0)], **_i(loc, s))
    else:
        return ast.Compare(toks[0], [toks[1]], [toks[2]], **_i(loc, s))

def do_stmt(s, loc, toks):
    if len(toks) == 3 and toks[1] == ':=':
        toks[0].ctx = ast.Store()
        return [ast.Assign([toks[0],], toks[2], **_i(loc, s))]
    elif toks[0] == 'if':
        return [ast.If(toks[1], _list(toks[3]), \
                _list(toks[5]) if len(toks) == 6 else [], **_i(loc, s))]
    elif toks[0] == 'while':
        return [ast.While(toks[1], _list(toks[3]), [], **_i(loc, s))]
    elif toks[0] == 'write':
        return [ast.Print(None, toks[1:], True, **_i(loc, s))]
    elif toks[0] == 'return':
        return [ast.Return(toks[1], **_i(loc, s))]
    elif isinstance(toks[0], ast.Call):
        return [ast.Expr(toks[0], **_i(loc, s))]
    else:
        return toks

def do_vardecl(s, loc, toks):
    return []

def do_body(s, loc, toks):
    return [toks[1:-1],]   # ignore keyword `begin` and `end`

def do_func(s, loc, toks):
    for t in toks[2]:
        t.ctx = ast.Param()
    args = ast.arguments(args=list(toks[2]), vararg=None, kwarg=None,
                         defaults=[], kw_defaults=[])
    return [ast.FunctionDef(toks[1].id, args, toks[3], [], **_i(loc, s))] + toks[4:]

def do_block(s, loc, toks):
    return [toks[:-1] + toks[-1]]

def do_program(s, loc, toks):
    return ast.Module(body=toks[2], **_i(loc, s))

def _dis(name, s, loc, toks):
    print '>', name, toks
    rc = globals()[name](s, loc, toks)
    print '<', name, rc
    return rc

for k in locals().keys():
    if k.startswith('do_'):
        name = k[3:]
        element = vars()[name]
        action = vars()[k]
        element.setName(name)
        #import functools
        #element.setParseAction(functools.partial(_dis, k))
        element.setParseAction(action)
        #expr.setDebug()
    
def compile(source, fname):
    """Compile PL/0 source into PyCodeObject"""
    try:
        a = program.parseString(source)[0]
    except ParseException, e:
        # see Objects/exceptions.c:SyntaxError_init
        raise SyntaxError(e.msg, (fname, e.lineno, e.col, e.line))

    return __builtins__.compile(a, fname, 'exec')

if __name__ == '__main__':
    tests = [
"""
{comment1}
program main;
var i, j, max, num;
begin
    i := 0; max := 100;
    num := 0;
    while i <= max do
    begin
        j := 2;
        while j < i do
            {comment2}
            if ( i-i/j*j ) = 0 then
                j := i+1
            else
                j := j+1;
        if j = i then
        begin
            write( i );
            num := num +1
        end;
        i := i+1
    end
end.
""",

"""
program main;
const _i=1, i_=2, _=3, __=4, s='hello';
begin
    write(_i, i_, _, __, s + ", world!")
end.
""",

"""
program main;
function add(a, b);
begin
    return a + b
end;

function procedure();
begin
    write('this is a procedure')
end

begin
    write('1 + 1 =', add(1, 2));
    { call python builtin }
    write('min(max(1, 2), 0) =', min(max(1, 2), 0));
    procedure()
end.
""",

"""
{ closure }
program main;
function rec(n);
    function w(n);
    begin
        if n <> 0 then
        begin
            w(n - 1);
            write(n)
        end
    end
begin write(w(n)) end

begin rec(10) end.
"""]

    for t in tests:
        print 'INPUT:'
        print t
        print 'OUTPUT:'
        dct = {}
        exec(compile(t, '<none>'), dct)

    from astpp import dump
    import pdb
    #a = program.parseString(t3)[0]
    #pdb.set_trace()
    #print dump(a)
    #print ast.dump(a, True, True)
    #exec(__builtins__.compile(a, "<ast>", "exec"))
