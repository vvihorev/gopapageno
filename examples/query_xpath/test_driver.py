import sys
import os
import subprocess

queries = [
    '/a',
    '\\\\a',
    '//a',
    '\\\\\\\\a',
    '/a/b',
    '\\\\a/b',
    '//a\\\\b',
    '\\\\\\\\a/b',
    '/a/b[/@c]',
    '/a/b[/@c="d"]',
    '/a/b[/@c and /@d]',
    '/a/b[/@c and /@d="e"]',
    '/a/b[/@c and /d\\\\e]',
]

for query in queries:
    with open('./tmp', 'w') as f:
        f.write(query)

    proc = subprocess.run(['./query_xpath', '-f', './tmp'], capture_output=True)
    if proc.returncode:
        print('TEST FAILED:', query, proc.stdout.decode())
        os.remove('./tmp')
        exit(1)
    else:
        print('PASSED:', query)
        if len(sys.argv) > 1:  # verbose test output
            print(proc.stdout.decode())
            input()
