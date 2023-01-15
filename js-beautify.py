#!/usr/bin/env python3
import os
import jsbeautifier


for cur, dirnames, filenames in os.walk('web/data/static/js/'):
    if 'vendor' in cur:
        continue

    for file in filenames:
        if file.endswith('.js'):
            filePath = os.path.join(cur, file)

            print('{}'.format(filePath))
            beauty = jsbeautifier.beautify_file(filePath)

            with open(filePath, 'bw') as f:
                f.write(beauty.encode('utf-8'))
