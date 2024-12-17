package nsenter

/*
#include <errno.h>
#include <fcntl.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

__attribute__((constructor)) void enter_namespace(void) {
    char *mydocker_pid = getenv("MYDOCKER_PID");
    if (mydocker_pid == NULL) {
        fprintf(stderr, "MYDOCKER_PID not found\n");
        return;
    }

    char *mydocker_cmd = getenv("MYDOCKER_CMD");
    if (mydocker_cmd == NULL) {
        fprintf(stderr, "MYDOCKER_CMD not found\n");
        return;
    }

    int i;
    char nspath[100];
    char *namespaces[] = {"ipc", "uts", "net", "pid", "mnt"}; // TODO: user namespace

    for (i = 0; i < 5; i++) {
        sprintf(nspath, "/proc/%s/ns/%s", mydocker_pid, namespaces[i]);
        int fd = open(nspath, O_RDONLY);
        if (setns(fd, 0) == -1) {
            fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
        }
        close(fd);
    }

    int res = system(mydocker_cmd);
    if (res == -1) {
        fprintf(stderr, "exec command failed: %s\n", strerror(errno));
    }
    exit(0);
}
*/
import "C"
