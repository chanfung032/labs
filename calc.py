import re

SPLIT_RE = re.compile(r'\d+|[+*/()-]')
OP2PRI = {"+": 0, "-": 0, "*": 1, "/": 1}

# 中缀转后缀
def expr2post(expr):
    post = []
    stack = []

    for e in SPLIT_RE.findall(expr):
        if e in "+-*/":
            while stack and stack[-1] in OP2PRI \
                and OP2PRI[stack[-1]] >= OP2PRI[e]:
                    post.append(stack.pop())
            stack.append(e)
        elif e == "(":
            stack.append(e)
        elif e == ")":
            while stack[-1] != '(':
                post.append(stack.pop())
            stack.pop()
        else:
            post.append(e)
    while stack:
        post.append(stack.pop())
    return post

import operator
OP = {
    '+': operator.add,
    '-': operator.sub,
    '*': operator.mul,
    '/': lambda a, b: int(float(a)/b),
}

def evalpost(post):
    stack = []
    for t in post:
	if t in OP:
	    r = stack.pop()
	    l = stack.pop()
	    stack.append(OP[t](l, r))
	else:
	    stack.append(int(t))
    return stack[0]

if __name__ == '__main__':
    print evalpost(expr2post("9+(3-1)*3+10/2"))