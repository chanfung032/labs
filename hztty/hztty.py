#!/usr/bin/env python
#-*-coding: utf8 -*-

"""
Reader for HZ-encoded text on alt.chinese.text forum.
About HZ-coding data format, please refer to rfc1843.txt.
"""

import sys
import struct
import locale

input = open(sys.argv[1], 'rb')

ASCII, ESCAPE, GB1, GB2 = range(4)

default_enc = locale.getpreferredencoding()

for line in input.readlines():
    buf = []
    mode = ASCII
    pos = 0
    line = line.rstrip()

    for c in line:
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
                buf.append(c)
                mode = ASCII
            else:
                print line
                raise ValueError("invalid character at pos %d" % pos)
        elif mode == GB1:
            if c == '~':
                mode = ESCAPE
            else:
                buf.append(struct.pack('B', 0x80 | ord(c)))
                mode = GB2
        elif mode == GB2:
            buf.append(struct.pack('B', 0x80 | ord(c)))
            mode = GB1

        pos = pos + 1

    try:
        gb_text = ''.join(buf)
        output = gb_text.decode('gb2312').encode(default_enc)
    except:
        print line
        raise

    print output

