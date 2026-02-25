#include <stdio.h>
#include <stdlib.h>
#include <errno.h>
#include <string.h>
#include <unistd.h>

void open_url(char *url) {
    char launch[255];
    sprintf(launch, "cmd /c start %s", url);
    system(launch); // Windows
    sprintf(launch, "x-www-browser '%s' &", url);
    system(launch); // Linux
    sprintf(launch, "open '%s'", url);
    system(launch); // MacOS
}

void error(char *str) {
    puts(str);
    puts(strerror(errno));
}

int main(int argc, char *argv[]) {
    char *phrase = argv[1];
    char *vars[] = {"RSS_FEED=http://feeds.feedburner.com/tecnoblog", NULL};
    int fd[2];
    if (pipe(fd) == -1) {
        error("Não foi possível criar o pipe");
    };
    pid_t pid = fork();
    if (pid == -1) {
        error("Não foi possível forkar o processo");

    }
    if (!pid) {
        dup2(fd[1], 1);
        close(fd[0]);
        if (execle("/usr/bin/python", "/usr/bin/python", "../rssgossip.py", "-u", phrase, NULL, vars) == -1) {
            error("Não foi possível rodar o programa");
        }
    }
    dup2(fd[0], 0);
    close(fd[1]);
    char line[255];
    while (fgets(line, 255, stdin)) {
        if(line[0] == '\t')
            open_url(line + 1);
    }
    return 0;
}
