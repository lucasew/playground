#include<stdio.h>
#include<stdlib.h>
#include<errno.h>
#include<fcntl.h>
#include<signal.h>
#include<string.h>
#include<sys/mman.h>
#include<sys/types.h>
#include<sys/wait.h>
#include<unistd.h>

#define fail(format, ...) { \
    fprintf(stderr, "%s:%i: ", __FILE__, __LINE__); \
    fprintf(stderr, format, ## __VA_ARGS__); \
    fprintf(stderr, ": %s\n", strerror(errno)); \
    if (child_id) { kill(child_id, 9); }; \
    exit(1); \
}

pid_t child_id = 0;
size_t buf_size = 0;
char* buf = NULL;
int exit_status = 0;

char stop = 0;

static void sigaction_handler(int sig) {
    pid_t pid;
    int status;

    while ((pid = waitpid(-1, &status, WNOHANG)) > 0) {
        if (pid == child_id) {
            printf("child killed!\n");
            stop = pid;
            exit_status = status;
        }
    }
}

int is_valid_fd(int fd) {
    return fcntl(fd, F_GETFL) != -1 || errno != EBADF;
}

int nonblock_fd(int fd) {
    return fcntl(fd, F_SETFL, fcntl(fd, F_GETFL) | O_NONBLOCK);
}

ssize_t handle_fd(int from_fd, int to_fd) {
    if (stop) {
        return 0;
    }
    ssize_t ret = read(from_fd, buf, buf_size);
    if (ret == -1 && errno == EAGAIN) {
        return 0;
    }
    if (ret == -1) {
        fail("can't read from fd %i", from_fd);
    }
    if (buf[0] == EOF) {
        printf("EOF %i\n", from_fd);
    }
    ssize_t wret = write(to_fd, buf, ret);
    if (wret == -1) {
        fail("can't write to fd %i", to_fd);
    }
    return wret;
}

int main(int argc, char *argv[]) {
    if (argc < 3) {
        fail("usage: tag command ...args");
    }
    buf_size = sysconf(_SC_PAGESIZE);
    if (buf_size == -1) {
        fail("can't get system page size");
    }
    printf("page size: %li\n", buf_size);
    buf = mmap(NULL, buf_size, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0);
    if (buf == MAP_FAILED) {
        fail("can't allocate the buffer");
    }

    /* fail("TODO: comment me later"); */
    char** args = argv;  // don't care about the first arg
    args++;
    char* tag = *args; // prefix for each line
    args++;
    char* executable = *args; // executable is first arg
    /* args++; // now args only has the args themselves */
    printf("tag: %s\n", tag);
    printf("executable: %s\n", executable);
    for (int i = 0; i < argc; i++) {
        if (args[i] == NULL) break;
        printf("arg: %s\n", args[i]);
    }
    /* printf("executable: %s\n", executable); */

    int stdin_pipe[2];
    int stdout_pipe[2];
    int stderr_pipe[2];
    if (pipe(stdin_pipe) == -1)
        fail("can't create pipe for stdin");
    if (pipe(stdout_pipe) == -1)
        fail("can't create pipe for stdout");
    if (pipe(stderr_pipe) == -1)
        fail("can't create pipe for stderr");
    child_id = fork();
    if (child_id == -1) {
        fail("can't fork")
    }

    if (child_id == 0) { // child
        close(stdin_pipe[1]);
        dup2(stdin_pipe[0], 0); // redirect stdin
        close(stdout_pipe[0]);
        dup2(stdout_pipe[1], 1); // redirect stdout
        close(stderr_pipe[0]);
        dup2(stderr_pipe[1], 2); // redirect stderr
        execvp(executable, args);
        fail("can't exec %s", executable);
    } else {
        close(stdin_pipe[0]);
        close(stdout_pipe[1]);
        close(stderr_pipe[1]);
        // nonblocking pipes so cooperative multitasking works
        char fcntlret = 0;
        if (nonblock_fd(0))
            fail("can't set stdin to nonblock");
        if (nonblock_fd(stdin_pipe[1]))
            fail("can't set stdin pipe to nonblock");
        if (nonblock_fd(stdout_pipe[0]))
            fail("can't set stdout pipe to nonblock");
        if (nonblock_fd(stderr_pipe[0]))
            fail("can't set stderr pipe to nonblock");

        while(!stop) {
            char bytes = 0;
            bytes += handle_fd(0, stdin_pipe[1]); // stdin is input, others are output
            bytes += handle_fd(stdout_pipe[0], 1);
            bytes += handle_fd(stderr_pipe[0], 2);
            if (!bytes) // 0 bytes == allow computer to breathe a bit
                usleep(20000);
        }
    }
    return exit_status;
}

