#!/usr/bin/env python

from subprocess import Popen, PIPE
from whoosh.index import create_in
from whoosh.fields import Schema, ID, TEXT

AUTH_URL = 'https://auth.sinas3.com/v1.0'
AK = '<accesskey>'
SK = '<secretkey>'
BUCKETS = ['ai', 'jn', 'oz']

schema = Schema(bucket=ID(stored=True),
                title=TEXT(stored=True),
                author=TEXT(stored=True))
ix = create_in('index', schema, indexname='mobi')
writer = ix.writer()
u = unicode
for bucket in BUCKETS:
    cmd = 'swift -A%s -U%s -K%s list %s' % (AUTH_URL, AK, SK, bucket), 
    p = Popen(cmd, stdout=PIPE, shell=True)
    for line in p.stdout:
        author, title = line.strip().split('/')
        title = title.replace('.mobi', '')
        print '/%s/%s/%s' % (bucket, author, title)
        writer.add_document(bucket=u(bucket), title=u(title), author=u(author))
writer.commit()
