#!/usr/bin/env python3

import os
import sys
import json
import ssl
import argparse
import urllib.request
import urllib.error

FOLDER_UID = '23fzcEW23rsf2'
FOLDER_NAME = 'Harvest 2.0'
API_TIMEOUT = 5

def main():

    args = read_args()

    dir_path = os.path.join(args.path, 'grafana/', args.directory)

    if not os.path.exists(dir_path):
        print('No dashboards for [{}]'.format(os.directory))
        os.exit(1)

    if not args.api_token:
        args.api_token = input('Enter Grafana API key:\n')

    headers = {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
        'Authorization': 'Bearer {}'.format(args.api_token) }

    url_prefix = '{}://{}{}'.format(
        'https' if args.https else 'http',
        args.addr,
        ':' + str(args.port) if args.port else '',
        )

    """
    # Create folder if doesn't exist
    print('Checking folder [{}]... '.format(FOLDER_NAME), end='')

    request = urllib.request.Request('{}/api/folder/{}'.format(url_prefix, FOLDER_UID))
    request.headers = headers

    exists = True
    
    try:
        r = urllib.request.urlopen(request, 
                data = None, 
                context = ssl._create_unverified_context(), 
                timeout = API_TIMEOUT )
    except urllib.error.HTTPError as err:
        if err.code == 404:
            exists = False
        else:
            print(err)
            sys.exit(1)

    if exists:
        print('exists')
    else:
        print('does not exist')
        print('Creating [{}]... '.format(FOLDER_NAME), end='')

        data = {}
        data['uid'] = FOLDER_UID
        data['title'] = FOLDER_NAME

        request = urllib.request.Request('{}/api/folders'.format(url_prefix), data = json.dumps(data).encode(), method = 'POST')
        request.headers = headers

        try:
           r = urllib.request.urlopen(request, 
                    context = ssl._create_unverified_context(), 
                    timeout = API_TIMEOUT )
    
        except urllib.error.HTTPError as err:
            print(err)
            try:
                print('response ({}): [{}] {}'.format(r.code, r.msg, r.reason))
            except:
                pass
            print('url was: [{}]'.format(request.full_url))
            sys.exit(1)

        else:
            if r.status == 200:
                print('DONE')
            else:
                print('failed')
                print('Grafana response ({}): [{}] {}'.format(r.code, r.msg, r.reason))
                sys.exit(1)
    """

    print('Importing/updating dashboards...')
    
    request = urllib.request.Request('{}/api/dashboards/db'.format(url_prefix))
    request.headers = headers


    for f in os.listdir(dir_path):

        if not f.endswith('.json'):
            continue

        fp = os.path.join(dir_path, f)

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
            print('Import success: https://{}{}'.format(url_prefix, resp['url']))


def read_args():

    p = argparse.ArgumentParser(
    description ='NetApp Harvest: Grafana API Utility')

    p.add_argument('-a', '--addr', 
        help        = 'Address of Grafana server (IP or hostname)',
        type        = str,
        dest        = 'addr',
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

    if len(sys.argv) > 1 and sys.argv[1] == 'help':
        p.print_help()
        sys.exit(0)

    args = p.parse_args()

    args.path = os.getenv('HARVEST_CONF', '/etc/harvest')

    args.addr.replace('https://', '')
    args.addr.replace('http://', '')
    
    return args


if __name__ == '__main__':
    main()
