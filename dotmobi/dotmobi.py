import os
import time

from flask import Flask, g, render_template

app = Flask(__name__)
app.debug = True

@app.before_request
def record_start_time():
    g.start_time = time.time()

@app.context_processor
def inject_render_time():
    return dict(render_time=time.time() - g.start_time)

@app.route('/')
def index():
    return render_template('index.html') 

from whoosh.index import open_dir
ix = open_dir('index', 'mobi', readonly=True)

from whoosh.query import Or
from whoosh.qparser import QueryParser

from whoosh.highlight import WholeFragmenter
fragmenter = WholeFragmenter()

@app.route('/search/<name>')
def search(name):
    with ix.searcher() as searcher:
        query = Or([QueryParser("title", ix.schema).parse(name),
                    QueryParser("author", ix.schema).parse(name)])
        results = searcher.search(query, limit=100)
        results.fragmenter = fragmenter
        return render_template('index.html', results=results)

from sae.const import APP_NAME
def book_url(hit):
    return 'http://%s-%s.stor.sinaapp.com/%s/%s.mobi' % \
            (APP_NAME, hit['bucket'], hit['author'], hit['title'])

app.jinja_env.globals.update(book_url=book_url)
