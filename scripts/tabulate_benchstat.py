"""Tabulate Benchstat

Tabulates benchstat output, puts thread count as columns, and query type as rows.
"""
import sys


def print_usage():
    print('Usage: tabulate_benchstat bench_output.txt')
    sys.exit(1)


if __name__ == '__main__':
    if len(sys.argv) < 2:
        print_usage()

    with open(sys.argv[1], 'r') as f:
        lines = [line.strip() for line in f.readlines() if line.startswith('Parse')]

    cells = {}
    for line in lines[:80]:
        axis = line.split()[0].split('/')[3:5]
        axis = [ax.split('=')[1] for ax in axis]
        query, goroutine = axis
        cells[(query, goroutine)] = line.split()[-3]

    queries = [f'A{i}' for i in range(1, 9)] + [f'B{i}' for i in range(1, 3)]
    goroutines = [f'{i}' for i in range(1, 9)]

    print('\t'.join([''] + [goroutine for goroutine in goroutines]))
    for query in queries:
        print('\t'.join([query] + [cells[(query, goroutine)] for goroutine in goroutines]))
