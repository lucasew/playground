#include <stdio.h>
#include <stdlib.h>
#include "ikkp_server.h"
#include "error.h"
#include <signal.h>
#include <unistd.h>
#include <strings.h>

int listener_d;

void handle_shutdown(int sig) {
    if(listener_d)
        ikkp_close(listener_d);
    fprintf(stderr, "Bye!\n");
    exit(0);
}

int catch_signal(int sig, void (*handler)(int)) {
    struct sigaction action;
    action.sa_handler = handler;
    sigemptyset(&action.sa_mask);
    action.sa_flags = 0;
    return sigaction(sig, &action, NULL);
}

int main(int argc, char *argv[]) {
    if(catch_signal(SIGINT, handle_shutdown) == -1)
        error("Não foi possível setar o handler de interrupção: ");
    listener_d = ikkp_listen(30000);
    if (listener_d == -1)
        error("Não foi possível escutar a porta: ");
    char buf[255];
    printf("Escutando por conexões...");
    while(1) {
        int connect_d = ikkp_accept(listener_d);
        if (connect_d == -1)
            error("Não foi possível aceitar a conexão: ");
        if(ikkp_recv(connect_d, buf, sizeof(buf)) == -1)
            error("Não foi possível ler a conexão: ");
        if (strncasecmp("Who's there?", buf, 12))
            ikkp_send(connect_d, "You should say 'Who's there?'!");
        else {
            if(ikkp_send(connect_d, "Oscar\r\n") != -1) {
                ikkp_recv(connect_d, buf, sizeof(buf));
                if(strncasecmp("Oscar who?", buf, 10))
                    ikkp_send(connect_d, "You should say 'Oscar who?'!\r\n");
                else
                    ikkp_send(connect_d, "Oscar silly question you get a silly answer \r\n");
            }
        }
        ikkp_close(connect_d);
    }
    return 0;
}
