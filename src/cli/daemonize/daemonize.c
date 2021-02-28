#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <syslog.h>

int daemonize(char *bin, char *args[]) {

    // open syslog to send messages
    openlog("harvest daemonize", LOG_PID, LOG_USER);
    syslog(LOG_INFO, "entrypoint");

    if (fork() != 0)
        return 0; // parent exits
    // child continues...
    syslog(LOG_INFO, "after first fork");

    // get new session ID
    if (setsid() == -1)
        return -1;

    // second fork, so we are not session leader
    if (fork() != 0)
        return 0;
    // (grand) child continues
    syslog(LOG_INFO, "after second fork");

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

    // finally execv so the poller has its own cmdline identity
    syslog(LOG_INFO, "before execv");
    execv(bin, args);
    syslog(LOG_INFO, "after execv, oops!");
    return -1; // if we got here, something went wrong
}


int main(int argc, char *argv[]) {

    if (argc == 1 || argv[1] == "-h" || argv[1] == "--help") {
        printf("Usage: ./daemonize <program path> <arg vector>\n");
        exit(0);
    }

    // construct arg vector, so we can detect errors more easily
    char *Args[argc-1];
    char ArgsFlat[200];
    int i;

    //ArgsFlat = "";

    // null terminated vector
    Args[0] = strrchr(argv[1], '/');
    if (Args[0] == NULL)
        Args[0] = argv[1];
    else
        Args[0]++;

    for (i=2; i<argc; i++)
        Args[i-1] = argv[i];
    Args[i-1] = NULL;

    for (i=0; i<argc-1; i++)
        strcat(ArgsFlat, Args[i]);

    printf("Daemonizing (%s) [%s]\n", argv[1], ArgsFlat);


    exit( daemonize(argv[1], Args) );
}