#!/usr/bin/env python

"""Paper Cropper for Kindle
"""

import copy
from pyPdf import PdfFileWriter, PdfFileReader

def convert(filename):
    inp = PdfFileReader(open(filename, 'rb'))
    outp = PdfFileWriter()

    for page in inp.pages:
        page1 = copy.copy(page)
        page2 = copy.copy(page)

        UL = page.mediaBox.upperLeft
        UR = page.mediaBox.upperRight
        LL = page.mediaBox.lowerLeft
        LR = page.mediaBox.lowerRight

        # left column
        page1.mediaBox.upperLeft = (UL[0], UL[1])
        page1.mediaBox.upperRight = (UR[0]/2, UR[1])
        page1.mediaBox.lowerLeft = (LL[0], LL[1])
        page1.mediaBox.lowerRight = (LR[0]/2, LR[1])
        outp.addPage(page1)

        # right column
        page2.mediaBox.upperLeft = (UR[0]/2, UL[1])
        page2.mediaBox.upperRight = (UR[0], UR[1])
        page2.mediaBox.lowerLeft = (LR[0]/2, LR[1])
        page2.mediaBox.lowerRight = (LR[0], LR[1])
        outp.addPage(page2)

    outp.write(open(filename+'.2', 'wb'))

if __name__ == '__main__':
    import sys
    if len(sys.argv) < 2:
        print 'Usage: k2pdf.py file [file ...]'
        sys.exit(0)

    for filename in sys.argv[1:]:
        convert(filename)
