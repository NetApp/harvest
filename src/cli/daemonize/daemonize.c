#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <syslog.h>
#include <errno.h>

/*
    Executes command as a daemon process. Unlike what 
    is done usually (two forks and exit), we perform
    execv after second fork, such that the daemon has
    it's own cmdline identity. This is helpful to
    trace status of pollers.

    Script is only intended to daemonize Harvest pollers.
    Daemonizing other processes should be ok, as long as
    the command line arguments are correct, but no
    guarantee.

    The daemonize function is implemented in the quick
    and dirty way: some system calls would need better
    error-checking for a more safe implementation.

    Usage:
        ./daemonize <executable> [args...]

    Arguments:
        - executable    path to executable program
        - args          optional, passed to daemon unchanged
*/


int daemonize(char *bin, char *args[]) {

    // open syslog to send messages
    openlog("harvest daemonize", LOG_PID, LOG_USER);

    if (fork() != 0)
        return 0; // parent exits
    // child continues...

    // get new session ID
    if (setsid() == -1) {
        syslog(LOG_ERR, "setsid: %s", strerror(errno));
        return -1;
    }

    // second fork, so we are not session leader
    if (fork() != 0)
        return 0;
    // (grand) child continues

    // clean file permissions
    umask(0);
    // change working directory to root
    chdir("/");

    // close FDs, if choose reasonable number
    int maxfds, fd;
    if ((maxfds = sysconf(_SC_OPEN_MAX)) == -1)
        maxfds = 256;
    for (fd=0; fd<maxfds; fd++)
        close(fd);

    // forwards standard FDs to devnull
    close(STDIN_FILENO);
    fd = open("/dev/null", O_RDWR);
    dup2(STDIN_FILENO, STDOUT_FILENO);
    dup2(STDIN_FILENO, STDERR_FILENO);

    // finally execv so the daemon has its own cmdline identity
    execv(bin, args);
     // if we got here, something went wrong
    syslog(LOG_ERR, "execv: %s", strerror(errno));
    return -1;
}

int main(int argc, char* argv[]) {

    // at least one argument is required
    if (argc < 2 || argv[1] == "-h" || argv[1] == "--help") {
        printf("Usage: ./daemonize <executable> [args...]\n");
        exit(0);
    }

    // construct path to executable and arg vector
    // this is done here, so errors are detected early on
    char* Args[100];
    char path[100];
    int i;

    strcpy(path, argv[1]);

    Args[0] = "poller";

    for (i=1; i<argc-1; i++) {
        Args[i] = argv[i+1];
    }

    // arg vector should be null-terminated
    Args[i] = NULL;

    return daemonize(path, Args);
}
