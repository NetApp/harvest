#!/usr/bin/env python3

import os
import json
import ssl
import argparse
import urllib.request
import urllib.error


def main():

    args = read_args()

    if not args.api_token:
        args.api_token = input('Enter Grafana API key:\n')


    request = urllib.request.Request('{}://{}{}/api/dashboards/db'.format(
        'https' if args.https else 'http',
        args.url,
        ':' + str(args.port) if args.port else '',
        ))

    request.headers = {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
        'Authorization': 'Bearer {}'.format(args.api_token) }

    print('******')
    print(request.full_url)
    print('******')

    for f in os.listdir('{}/{}'.format(args.path, args.directory)):

        if not f.endswith('.json'):
            continue

        fp = '{}/{}/{}'.format(args.path, args.directory, f)

        with open(fp) as fd:
            try:
                dashboard = json.load(fd)
            except json.decoder.JSONDecodeError as ex:
                print('Error reading [{}]: {}'.format(f, ex))
                continue

        dashboard['id'] = None
        data = {}
        data['dashboard'] = dashboard
        data['overwrite'] = True

        try:
            r = urllib.request.urlopen(request, 
                data = json.dumps(data).encode(),
                context = ssl._create_unverified_context(), 
                timeout = 2 )
        except (urllib.error.URLError, urllib.error.HTTPError) as ex:
            print('Importing [{}] failed, API error: {}'.format(f, ex))
            print('Requested URL: [{}]'.format(request.full_url))
            headers = request.headers
            if 'Authorization' in headers:
                headers['Authorization'] = '*********************'
            print('Headers: [{}]\n\n'.format(headers))
            raise

        if r.status != 200:
            print('Importing [{}] failed: [{}] [{}] {}'.format(f, r.code, r.msg, r.reason))
        else:
            resp = json.loads(r.read().decode())
            print('Import success: https://{}{}'.format(args.url, resp['url']))


def read_args():

    p = argparse.ArgumentParser(
    description ='NetApp Harvest: Grafana API Utility')

    p.add_argument('-u', '--url', 
        help        = 'Grafana server hostname or IPv4',
        type        = str,
        dest        = 'url',
        required    = True,
        )

    p.add_argument('-p', '--port', 
        help        = 'Grafana server port',
        type        = int,  
        dest        = 'port',
        default     = None,
        )

    p.add_argument('-t', '--api_token',
        help        = 'Grafana API token',
        type        = str,
        dest        = 'api_token',
        required    = False,
        default     = None  
        )

    p.add_argument('-d', '--directory',
        help        = 'Directory from which to import dashboards',
        type        = str,
        dest        = 'directory',
        required    = True
        )
    p.add_argument('--https',
        help        = 'Force to use HTTPS',
        type        = bool,
        dest        = 'https',
        default     = False
        )        

    args = p.parse_args()

    args.path = os.path.dirname(os.path.abspath(__file__))
    if args.path.endswith('utils'):
        args.path = args.path.replace('utils', '')
    if args.path.endswith('utils/'):
        args.path = args.path.replace('utils/', '')
    if args.path.endswith('/'):
        args.path = args.path.rstrip('/')
    if args.directory.endswith('/'):          
        args.directory = args.directory.rstrip('/')
    
    return args


if __name__ == '__main__':
    main()
