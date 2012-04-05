#!/usr/bin/env python
#-*-coding: utf8 -*-

"""
Reader for HZ-encoded text on alt.chinese.text forum.
About HZ-coding data format, please refer to rfc1843.txt.
"""

import sys
import struct
import locale

ASCII, ESCAPE, GB1, GB2 = range(4)

input = open(sys.argv[1], 'rb').read()

buf = []
mode = ASCII
for c in input:
    if mode == ASCII:
        if c == '~':
            mode = ESCAPE
        else:
            buf.append(c)
    elif mode == ESCAPE:
        if c == '{':
            mode = GB1
        elif c == '}':
            mode = ASCII
        elif c == '~':
            buf.append(c)
            mode = ASCII
        elif c == '\n':
            mode = ASCII
        else:
            raise ValueError("invalid character sequence")
    elif mode == GB1:
        if c == '~':
            mode = ESCAPE
        else:
            buf.append(struct.pack('B', 0x80 | ord(c)))
            mode = GB2
    elif mode == GB2:
        buf.append(struct.pack('B', 0x80 | ord(c)))
        mode = GB1

gb_text = ''.join(buf)
output = gb_text.decode('gb2312').encode(locale.getpreferredencoding())
print output

