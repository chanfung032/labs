# vi: filetype=python

import sys, os.path
sys.path.insert(0, os.path.abspath('site-packages'))

# `whoosh/filedb/compound.py` needs to import `mmap` module,
# but sae python does not have `mmap` module, here we fake to
# make it importable
import imp
m = imp.new_module('mmap')
def _mmap(*args, **kws):
    raise OSError()
m.mmap = _mmap
m.ACCESS_READ = 1
sys.modules['mmap'] = m

from dotmobi import app as application
