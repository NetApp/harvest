#!/usr/bin/env python3

"""

NetApp Harvest 2.0: the swiss-army-knife for datacenter monitoring

This module is a CLI wrapper for starting and monitoring your 
pollers. By default pollers are started as daemon processes, but
you can also start one in foreground, which is handy for debugging.

Author & Maintainer: Vachagan Gratian
Project Manager: Georg Mey
Contact: ng-harvest-maintainers@netapp.com

This project is based on NetApp Harvest, authored by 
Chris Madden in 2016.

Copyright (c) 2020 NetApp GmbH

"""

import os
import sys
import time
import signal
import syslog
import argparse
import yaml
import shutil
import subprocess

BOLD = '\033[1m' 
END = '\033[0m'


def main():

    args = read_args()

    #print("args=", args)

    #if args.version:
    #    print('NetApp Harvest {}'.format(__version__))
    #    return

    if os.geteuid() == 0:
        print('Warning: Running Pollers as a priviliged user is ' \
            'not safe and will be disabled in future releases\n')

    # For other actions, we need list of pollers from config
    all_pollers = get_poller_names(args.path, args.config)

    # If user didn't specify pollers, take all pollers from config
    if not args.pollers:
        pollers = all_pollers
    # Verify validity of poller names from input
    else:
        pollers = []
        for p,dc in all_pollers:
            if p in args.pollers:
                args.pollers.remove(p)
                pollers.append((p,dc))

        for p in args.pollers:
            print('Poller [{}] not defined in config'.format(p))

    # No need to continue if there are no valid pollers
    if not pollers:
        return

    args.level = 'DEBUG' if args.verbose else 'INFO'

    # Startup poller in foreground mode, but only list collectors 
    # and exporters and exit
    if args.action == 'test':
        if len(pollers) != 1:
            print('Choose exactly one poller to test')
            return
        print('Testing startup of collectors and exporters. ' \
            'This might take a few seconds')
        poller,__ = pollers.pop()
        args.level = 'CRITICAL'
        args.foreground = True
        args.debug = True
        args.test = True
        return start_poller(poller, args)

    # Starting a poller in foreground is a special case:
    # - we can only start one single poller at a time
    # - only in debug mode
    # This makes it save to run a poller in fg while it might be
    # also running as a daemon in bg, because we won't accidentally 
    # overwrite its PID file and exporters won't emit metrics to DBs.
    if args.foreground:
        if len(pollers) != 1:
            print('You can start only one poller in foreground mode')
            return
        if not args.debug:
            print('Setting debug mode ON, otherwise starting a ' \
                'poller in the foreground might not be safe and ' \
                'could corrupt PID files or DBs')
            #args.debug = True
        poller,__ = pollers.pop()
        return start_poller(poller, args)

    # From this point on we assume our pollers are daemons
    print()
    print('{:<20} {:<30} {:<20} {:<10}'.format('DATACENTER', 'POLLER', 'STATUS', 'PID'))
    print('++++++++++++++++++++ ++++++++++++++++++++++++++++++ ++++++++++++++++++++ ++++++++++')

    # Pollers will be asked to delay their startup incrementally
    # This is to prevent crashing the system with 100s of threads
    # and 100s of sockets starting simultaneously
    for poller in pollers:

        p, dc = poller

        # Each of the three methods called is expected to return
        # a tuple of two elements: status (str) and PID (int)
        if args.action == 'status':
            print('{:<20} {}{:<30}{} {:<20} {:<10}'.format(dc, BOLD, p, END, *get_status(p)))

        if args.action == 'stop' or args.action == 'restart':
            print('{:<20} {}{:<30}{} {:<20} {:<10}'.format(dc, BOLD, p, END, *stop_poller(p)))

        if args.action == 'start' or args.action == 'restart':
            print('{:<20} {}{:<30}{} {:<20} {:<10}'.format(
                dc, BOLD, p, END, *start_poller(p, args)))

    print('++++++++++++++++++++ ++++++++++++++++++++++++++++++ ++++++++++++++++++++ ++++++++++')
    return


def get_poller_names(path, filename):
    """
    Get a list of poller names from the configuration file.
    Exit if filename not found.

    Parameters
    ----------
    path :  string
        package (root) directory path
    filename : string
        the config file name

    Returns
    -------
    list of poller names
    """
    
    fp = '{}/{}'.format(path, filename)
    content = None

    try:
        with open(fp) as f:
            config = yaml.safe_load(f)
    except FileNotFoundError:
        print('Config file [{}] not found'.format(fp))
        sys.exit(0)
    except PermissionError:
        print('No permission to read config [{}]'.format(fp))
        sys.exit(0)
    except yaml.scanner.ScannerError as e:
        # This is a common issue when tabs are present in file
        # Try to fix it by replacing it with 2 or 4 spaces
        problem = getattr(e, 'problem', '')  # "found character '\\t' that cannot start any token"
        if 'found character \'\\t\'' in problem:
            print('Error loading yaml file, likely because tabs are ' \
                'mixed with whitespaces')
            with open(fp) as f:
                content = f.readlines()
        else:
            print('Failed to load yaml file, error from [pyyaml]: {}'.format(e))
            sys.exit(0)

    # Try to fix tab issues
    if content:
        print('Trying to fix by replacing tabes with 2 spaces... ', end='')
        fixed_content = [line.replace('\t', '  ') for line in content]

        try:
            config = yaml.safe_load('\n'.join(fixed_content))
        except yaml.parser.ParserError:
            print(' Failed')
            
            print('Trying to fix by replacing tabes with 4 spaces... ', end='')
            fixed_content = [line.replace('\t', '    ') for line in content]

            try:
                config = yaml.safe_load('\n'.join(fixed_content))
            except yaml.parser.ParserError:
                print(' Failed')
                print('Try to fix config file and try again.')
                fixed_content = None
            else:
                print('OK')

        else:
            print('OK')

        if fixed_content:
            if input('Overwrite changes? [y/N]: ') in ('y', 'Y', 'yes', 'Yes'):
                with open(fp, 'w') as f:
                    try:
                        f.writelines(fixed_content)
                    except:
                        raise  

    try:
        pollers = list(config['Pollers'].keys())
    except KeyError:
        print('Invalid config file: could not find "Pollers" section')
        sys.exit(0)

    dc = config.get('Defaults', {}).get('datacenter', '')

    return [(poller, config['Pollers'][poller].get('datacenter', dc)) for poller in pollers]


def get_status(poller_name):
    """
    
    Trace poller status. 
    
    This is partially guesswork. If we can't find its PID file, we
    assume poller is not running. If PID file exists, but process
    is not running, we assume it has crashed (since pollers have
    to clean up PID files on exit).
    
    Parameters
    ----------
    poller_name : string
        name of the poller

    Returns
    -------
    tuple of two elements:
        status (string), PID (int)

    """

    pidfp = 'var/.{}.pid'.format(poller_name)

    # Try to read poller PID file
    try:
        with open(pidfp) as f:
            pid = int( f.read() )
    # if no PID file, assume process is not running
    except FileNotFoundError:
        return 'NOT RUNNING', ''
    # Corrupt PID probably means poller crashed
    except ValueError:
        clean_pidf(pidfp)
        return 'CRASHED?', ''

    # Signal should raise exception if process is not running
    # without killing it
    try:
        os.kill(pid, 0)
    # Process exited without cleaning its PID file, so it must
    # have terminated unexpectedly
    except ProcessLookupError:
        clean_pidf(pidfp)
        return 'INTERRUPTED', pid

    # If we reached here the process with the PID is running,
    # just double check if the process matches the poller since
    # it s possible that PID was reassigned to another process
    try:
        with open('/proc/{}/cmdline'.format(pid)) as f:
            cmdline = f.read()
    except OSError as ex:
        print('Unable to read [/proc/{}/cmdline]: {}'.format(pid, ex))
        return 'FAILED TO CHECK', pid

    cmdargs = cmdline.replace('\x00', ' ').rstrip().split()

    if len(cmdargs) > 2 and cmdargs[2] == poller_name:
        return 'RUNNING', pid
    
    # Clean PID here? Most times that should be fine
    # But if this goes wrong we end up creating same
    # daemon over and over again
    return 'PID CHANGED?', pid


def filter_args(args):
    """

    Keep only options that are relevant to poller
    (i.e. can be accepted as CLI flags by the module 
    poller/poller.py)

    Parameters
    ----------
    args : ArgumentParser namespace object
        arguments to be passed to poller

    Returns
    -------
    dict 
    """
    ignore = (
        'action', 
        'pollers', 
        'foreground', 
        'verbose',
        'level'
        )

    return {k:v for k,v in vars(args).items() if k not in ignore}

def start_poller(poller_name, args):
    """

    Either directly start poller or start as daemon and 
    return its status.

    Parameters
    ----------
    poller_name : string
        name of the poller
    args : ArgumentParser namespace object
        arguments to be passed to poller

    Returns
    -------
    tuple of two elements:  
        status (string), pid (int)    
    
    """

    # Check that it's not already running
    status, pid = get_status(poller_name)
    if status == 'RUNNING':
        return 'ALREADY RUNNING', pid

    # Construct CMD arguments
    cmd_args = [os.path.join(args.path, 'bin/poller'), "poller"]
    cmd_args.append("-poller")
    cmd_args.append(poller_name)

    for k,v in filter_args(args).items():
        if type(v) is bool and v is True:
            cmd_args.append('-'+k)
        elif type(v) is list and len(v) > 1:
            cmd_args.append('-'+k)
            cmd_args += v
        elif type(v) is str and v != '':
            cmd_args.append('-'+k)
            cmd_args.append(v)
        elif type(v) is int:
            cmd_args.append('-'+k)
            cmd_args.append(str(v))

    if not args.foreground:
        cmd_args.append('-daemon')

    # Start in foreground
    if args.foreground:
        try:
            os.execv(cmd_args[0], cmd_args[1:])
        except:
            raise
        return

    # Start as daemon
    daemonize(poller_name, cmd_args, args.path)

    # Poller should immediately write its PID to file at startup
    # Allow for some delay and retry checking status a few times
    for retry in range(5):
        time.sleep(0.1)
        status, pid = get_status(poller_name)
        if pid:
            return status, pid
        elif status != 'NOT RUNNING':
            return status, pid
        else:
            # no pid means poller has not properly started (yet)
            pass
    return get_status(poller_name)


def stop_poller(poller_name):

    """
    Stop poller if it's running

    Parameters
    ----------
    poller_name : string
        name of the poller

    Returns
    -------
    tuple of two elements:
        status (string), pid (int) 

    """
    # Check poller status and identity before stopping it
    status, pid = get_status(poller_name)

    if status != 'RUNNING':
        return status, pid

    # Send terminate signal to process
    try:
        os.kill(pid, signal.SIGTERM)
    except ProcessLookupError:
        return 'ALREADY STOPPED', pid

    # Check that process actually stopped, wait up to 1s
    # since it might take some time to close threads
    # and cleanup
    for retry in range(5):
        time.sleep(0.2)
        try:
            os.kill(pid, 0)
        except ProcessLookupError:
            return 'STOPPED', pid
    return 'STOPPING FAILED', pid


def daemonize(poller_name, cmd, path):
    """
    Start poller as daemon process. Since we fork, all 
    error messages will be sent to syslog.

    Parameters
    ----------
    poller_name : string
        name of poller
    cmd : list
        arguments to be passed
    path : string
        harvest installation directory

    """

    try:
        if os.fork():
            return
    except OSError as ex:
        syslog.syslog(syslog.LOG_ERR, '[poller={}] Error during ' \
            'fork: {}'.format(poller_name, ex))
        raise

    if os.setsid() == -1:
        syslog.syslog(syslog.LOG_WARNING, '[poller={}] Creating ' \
            'session ID failed'.format(poller_name))
        # ignore and try anyway

    # Forward standard file descriptors to devnull and close
    devnull = os.open(os.devnull, os.O_RDWR)
    os.dup2(devnull, 0)
    os.dup2(devnull, 1)
    os.dup2(devnull, 2)
    os.close(devnull)

    # Set file permissions: read for all, write for group
    os.umask(0o27)
    # Set CWD to package root directory
    os.chdir('/')

    try:
        os.execv(cmd[0], cmd[1:])
    except Exception as ex:
        syslog.syslog(syslog.LOG_ERR, '[poller={}] Failed starting ' \
            'subprocess {}: {}'.format(poller_name, cmd, ex))
        raise
    
    syslog.syslog(syslog.LOG_NOTICE, '[poller={}] Launched daemon ' \
        'process with PID [{}]'.format(poller_name, p.pid))

    sys.exit(0)



def clean_pidf(pidf):
    """
    Delete PID file

    Parameters
    ----------
    pidf : string
        PID filepath

    """
    try:
        os.remove(pidf)
    except FileNotFoundError:
        pass
    except OSError as ex:
        raise ex


def read_args():
    """
    Read CLI arguments

    Returns
    -------
    args : ArgumentParser namespace object

    """

    p = argparse.ArgumentParser(
        description     = 'NetApp Harvest2 Manager)',
        formatter_class = argparse.RawTextHelpFormatter)

    p.add_argument('action', 
        type        = str, 
        choices     = ['start', 'status', 'stop', 'restart', 'test', ''],
        help        = 'Action to take:\n\n' \
                            '- "start", "restart" - by default create daemon pollers.\n\n' \
                            '- "stop", "status - only consider daemon pollers. ' \
                            '(a poller started in foreground (with "-f") runs in restricted ' \
                            'mode and should be stopped with Keyboard Interrupt (CTRL+C)).\n\n' \
                            '- "test" - will always start a poller in foreground, print status of ' \
                            'collectors and exporters and exit.\n\n',
        default     = '',
        nargs       = '?',
        )
    p.add_argument('pollers',
        help        = 'List of pollers for which you call the action, if left empty all pollers ' \
                    'are yielded',
        nargs       = '*', 
        default     = []
        )
    p.add_argument('-v', '--verbose',
        help        = 'Start poller(s) in verbose mode (pollers will log extensively)',
        action      = 'store_true',
        dest        = 'verbose',
        default     = False
        )
    p.add_argument('-d', '--debug',
        help        = 'Start poller(s) in debug mode (no data will be exported to DBs)',
        action      = 'store_true',
        dest        = 'debug',
        default     = False
        )
    p.add_argument('-f', '--foreground',
        help        = 'Start poller in foreground mode, implies debug (only one ' \
                        'single poller can be started in foreground mode)',
        action      = 'store_true',
        dest        = 'foreground',
        default     = False
        )      
    p.add_argument('-c', '--collectors', 
        help        = 'Only start these classes of collectors, e.g. "Zapi", ' \
                        '"ZapiPerf" (intended for debugging)',
        nargs       = '*',
        dest        = 'collectors',
        type        = str,
        default=[]
        )
    p.add_argument('-n', '--names', 
        help        = 'Only start collectors with these names, e.g. "WAFL", ' \
                        '"SystemNode" (intended for debugging)',
        nargs       = '*',
        dest        = 'names',
        type        = str,
        default     = []
        )               
    p.add_argument('--config', 
        help        = 'Configuration file (default: config.yaml)',
        type        = str,
        default     = 'config.yaml'
        )
    p.add_argument('--path', 
        help        = 'Harvest installation directory (default: "/opt/harvest2")',
        type        = str,
        default     = '/opt/harvest2'
        ) 
    p.add_argument('--loglevel', 
        help        = 'Logging level (0= Trace, 1= Debug, 2= Info)',
        type        = int, 
        default     = 2,
        ) 


    args = p.parse_args()

    if not args.action and not args.version:
        p.print_help()
        sys.exit(0)

    return args


if __name__ == '__main__':
    main()
