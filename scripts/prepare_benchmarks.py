"""Prepare Benchmarks

This scripts takes as input a number of Go benchmark files, possibly obtained by benchmarks in different packages,
and modifies them so that they can be used to compare them with utilities such as `benchstat`.
"""
import sys
import os


def print_usage():
    print('Usage: prepare_benchmarks file1 file2 ... fileN')
    sys.exit(1)


if __name__ == '__main__':
    if len(sys.argv) < 2:
        print_usage()

    strategies = ['copp', 'aopp', 'opp']

    for i, filename in enumerate(sys.argv[1:]):
        with open(filename) as f:
            lines = f.read()

            for strat in strategies:
                lines = lines.replace(strat, '')
                lines = lines.replace(strat.upper(), '')

        fn, ext = os.path.splitext(filename)
        with open(fn + "_stat" + ext, 'w') as f:
            f.write(lines)
