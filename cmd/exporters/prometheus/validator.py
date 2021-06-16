#!/usr/bin/env python3

import argparse
import regex
import signal
import sys
import time
import urllib.request

# error summary
errors = {
    'corrupt_metrics'       : 0,
    'corrupt_labels'        : 0,
    'corrupt_metatags'      : 0,
    'inconsistent_labels'   : 0,
    'duplicate_labels'      : 0,
    'missing_metatags'      : 0,
    'missing_newlines'       : 0,
    }

# cache label keys of seen metrics to check for consistency
label_cache = {}  # str -> set
# cache metrics for which we have seen metatags
help_cache = {}   # str -> bool
type_cache = {}   # str -> bool

# regular expressions to match metric
metric_pattern = regex.compile(r'^(\w+)\{(.+)\} \d+(\.\d+(e[-+]\d+)?)?$')
# pattern to match HELP/TYPE metatags
tag_pattern = regex.compile(r'^# (\w+) (\w+) .*$')
# label name must start with alphabetical char
# see: https://github.com/prometheus/common/blob/main/model/labels.go#L94
label_pattern = regex.compile(r'^([_a-zA-Z]\w*)="[^"]??"$', flags=regex.ASCII)

# tty colors
END = '\033[0m'
BOLD = '\033[1m'
RED = '\033[91m'
GREEN = '\033[92m'
YELLOW = '\033[93m'
PINK = '\033[95m'

def main():
    # parse arguments
    a = get_args()

    # make sure to print errors before exiting
    signal.signal(signal.SIGINT, terminate)

    # run the scrapes
    for i in range(a.scrapes):
        metrics = get_batch_metrics(a.addr, a.port)
        print('{}-> scrape #{:<4} - scraped metrics: {}{}'.format(BOLD, i+1, len(metrics), END))

        if not metrics.endswith('\n'):
            errors['missing_newlines'] += 1
            print('   {}missing newline at the end of metric batch{}'.format(PINK, END))

        for m in metrics.splitlines():

            # skip newline
            if m == '\n':
                continue

            # handle metatag
            if len(m) and m[0] == '#':
                ok, tag, metric_name = check_metatag(m)
                if not ok:
                    errors['corrupt_metatags'] += 1
                    print('   corrupt {} metatag:'.format(tag))
                    print('   [{}{}{}]'.format(RED, m, END))
                elif tag == 'HELP':
                    help_cache[metric_name] = True
                elif tag == 'TYPE':
                    type_cache[metric_name] = True
                continue

            # check general metric intergrity and parse raw labels substring
            ok, metric_name, raw_labels = check_metric(m)

            if not ok:
                errors['corrupt_metrics'] += 1
                print('   corrupt metric format:')
                print('   [{}{}{}]'.format(RED, m, END))
                continue

            # check labels integrity
            ok, labels = parse_labels(raw_labels) # list
            if not ok:
                errors['corrupt_metrics'] += 1
                print('   corrupt metric format (labels):')
                print('   [{}{}{}]'.format(RED, m, END))
                continue
            
            # check for duplicate labels
            duplicates = set([l for l in labels if labels.count(l) > 1])
            if duplicates:
                errors['duplicate_labels'] += 1
                print('   duplicate labels ({}):'.format(', '.join(duplicates)))
                print('   [{}{}{}]'.format(RED, m, END))

            labels = set(labels)

            # compare with cached labels for consistency
            cached_labels = label_cache.get(metric_name, None)
            if cached_labels == None:
                label_cache[metric_name] = labels
            else:
                missing = cached_labels - labels
                added = labels - cached_labels
                if missing or added:
                    errors['inconsistent_labels'] += 1
                    print('   inconsistent labels (cached: {}):'.format(' '.join(cached_labels)))
                    if missing:
                        print('     - missing ({})'.format(', '.join(missing)))
                    if added:
                        print('     - added ({})'.format(', '.join(added)))
                    print('   [{}{}{}]'.format(RED, m, END))

            # optionally check for metatags
            # each metrics should at least once include HELP/TYPE metametric
            if a.metatags:
                has_help = help_cache.get(metric_name, False)
                has_type = type_cache.get(metric_name, False)
                if not has_help or not has_type:
                    errors['missing_metatags'] += 1
                    print('   {}missing metatags for metric [{}]{}'.format(RED, metric_name, END))
                    if not has_help:
                        print('     - HELP tag not detected')
                    if not has_type:
                        print('     - TYPE tag not detected')

        # sleep until next scrape
        time.sleep(a.interval)

    print_errors()
    # DONE

# Scrape an HTTP endpoint and return data
def get_batch_metrics(addr: str, port: int) -> [str]:
    try:
        return urllib.request.urlopen('http://{}:{}/metrics'.format(addr, port)).read().decode()
    except urllib.error.URLError as err:
        print(err)
        return []

# validate metric format (without labels), extract name and labels substring
def check_metric(metric: str) -> (bool, str, str):
    match  = metric_pattern.match(metric)
    if match:
        try:
            return True, match.captures(1)[0], match.captures(2)[0]
        except Exception as ex:
            print('regex exception: {}'.format(ex))
    return False, '', ''

def check_metatag(metric: str) -> (bool, str, str):
    match = tag_pattern.match(metric)
    if match:
        try:
            return True, match.captures(1)[0], match.captures(2)[0]
        except Exception as ex:
            print('regex exception: {}'.format(ex))
    return False, '', ''

# parse label keys from raw labels substring
def parse_labels(labels: str) -> (bool, [str]):
    keys = []
    for pair in labels.split(','):
        match = label_pattern.match(pair)
        if not match:
            return False, keys
        keys.append(match.captures(1)[0])

    return True, keys

def terminate(signum, frame):
    print('\n{}-> terminating validation session{}'.format(YELLOW, END))
    print_errors()
    sys.exit()

def print_errors():
    print('~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~')
    print('-> {} unique metrics validated'.format(len(label_cache)))
    total = sum(errors.values())
    if total == 0:
        print('{}-> OK - no errors detected{}'.format(GREEN, END))
    else:
        print('{}-> FAIL - {} errors detected{}'.format(RED, total, END))

    for k, v in errors.items():
        print('{:<30} - {:>8}'.format(k, v))
    print('~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~')


# Parse CLI arguments
def get_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(
        formatter_class = argparse.RawTextHelpFormatter,
        description = """Open Metric Validator using an HTTP endpoint

SYNOPSIS:
    Run this tool specifying the port of the Prometheus exporter. Then,
    start a Harvest poller that will serve the metrics on the port.
    (Start tools first, so no metatags are missed).

VALIDATION:
    Tool will validate integrity of the rendered metrics:
        - metric format
        - label integrity
        - label consistency
        - label duplicates
        - HELP/TYPE metatags (optional)"""
        )
    p.add_argument('-a', '--addr',
        help = 'Address of the HTTP endpoint (default: localhost)',
        dest = 'addr',
        type = str,
        default = 'localhost'
        )
    p.add_argument('-p', '--port',
        help = 'Port of the HTTP endpoint',
        dest = 'port',
        type = int,
        required = True
        )
    p.add_argument('-i', '--interval',
        help = 'Interval between scrapes (in seconds, default: 60)',
        dest = 'interval',
        type = int,
        default = 60
        )
    p.add_argument('-s', '--scrapes',
        help = 'Number of scrapes to run (default: 5)',
        dest = 'scrapes',
        type = int,
        default = 5
        )
    p.add_argument('-m', '--metatags',
        help = 'Check TYPE/HELP metatags (default: false)',
        dest = 'metatags',
        action = 'store_true',
        default = False
        )
    return p.parse_args()

if __name__ == '__main__':
    main()
