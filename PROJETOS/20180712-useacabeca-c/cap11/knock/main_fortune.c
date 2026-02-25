#include <sys/socket.h>
#include <stdlib.h>
#include <stdio.h>
#include "ikkp_server.h"
#include "error.h"

int main(int argc, char *argv[]) {
    char *suff = "\r\n"; // suffix
    char *advice[] = {
        "Take smaller bites",
        "Go for the tight jeans. No they do NOT make you look fat.",
        "One word: inappropriate",
        "Just for today, be honest. Tell your boss what you *really* think",
        "You might want to rethink that haircut"
    };
    int listener_d = ikkp_listen(30000);
    if (listener_d == -1)
        error("Não foi possível escutar por conexões: ");
    puts("Aguardando conexão");
    int connect_d; // Pra que redeclarar?
    char *msg;
    int err; // Onde podemos jogar os possíveis erros
    while (1) {
        connect_d = ikkp_accept(listener_d);
        if (connect_d == -1)
            error("Não foi possível aceitar a conexão: ");
        msg = advice[rand() % 5]; // A princípio é 5
        err = ikkp_send(connect_d, msg);
        if (err == -1)
            error("Não foi possível escrever no socket: ");
        ikkp_send(connect_d, suff);
        err = ikkp_close(connect_d);
        if (err == -1)
            error("Não foi possível fechar a conexão: ");
    }
}
