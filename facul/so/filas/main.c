#include <stdio.h>
#include <stdlib.h>
#include <mqueue.h>
#include <string.h>
#include <stdarg.h>
#include <errno.h>
#include <signal.h>
#include <unistd.h>

#define die_with(msg, args...) fprintf(stderr, "\033[0;31merro\033[0m: "); fprintf(stderr, msg, ## args); exit(1);
#define die_with_errno() die_with("%s", strerror(errno))
#define info(msg, args...) fprintf(stderr, "\033[0;32minfo\033[0m: "); fprintf(stderr, msg, ## args);

const char* INPUT_QUEUE = "/inputq";
const char* OUTPUT_QUEUE = "/outputq";

typedef struct {
    int x;
    int y;
} Command;

int shuffle() {
    return random() % 100;
}

void handle_send() {
    int res;
    struct mq_attr attr = {
        .mq_maxmsg = 10,
        .mq_msgsize = sizeof(res),
        .mq_flags = 0
    };
    info("inicializando queues...\n");
    mqd_t output_queue = mq_open(OUTPUT_QUEUE, O_RDWR|O_CREAT, 0666, &attr); // inicializa sempre a queue que cria o arquivo primeiro
    if (output_queue < 0) {
        die_with_errno();
    }
    mqd_t input_queue = mq_open(INPUT_QUEUE, O_RDWR);
    if (input_queue < 0) {
        die_with_errno();
    }
    for (;;) {
        Command cmd = {
            .x = shuffle(),
            .y = shuffle()
        };
        info("enviando conta\n");
        if(mq_send(input_queue, (char*) &cmd, sizeof(cmd), 0) < 0) {
            die_with_errno();
        }
        info("enviado: %i + %i\n", cmd.x, cmd.y);
        info("recebendo resultado\n");
        if(mq_receive(output_queue, (char*) &res, sizeof(res), 0) < 0) {
            die_with_errno();
        }
        info("recebido: %i\n", res);
        sleep(1);
    }
}

void handle_recv() {
    Command cmd;
    int res;
    struct mq_attr attr = {
        .mq_maxmsg = 10,
        .mq_msgsize = sizeof(cmd),
        .mq_flags = 0
    };
    info("inicializando queues...\n");
    mqd_t input_queue = mq_open(INPUT_QUEUE, O_RDWR|O_CREAT, 0666, &attr);
    if (input_queue < 0) {
        die_with_errno();
    }
    mqd_t output_queue = mq_open(OUTPUT_QUEUE, O_RDWR);
    if (output_queue < 0) {
        die_with_errno();
    }
    for (;;) {
        info("recebendo conta...\n");
        if(mq_receive(input_queue, (char*) &cmd, sizeof(cmd), 0) < 0) {
            die_with_errno();
        }
        info("recebido: %i + %i\n", cmd.x, cmd.y);
        res = cmd.x + cmd.y;
        info("enviando resultado...\n");
        if(mq_send(output_queue, (char*) &res, sizeof(res), 0) < 0) {
            die_with_errno();
        }
        info("enviado: %i\n", res);
        sleep(1);
    }
}


int main(int argc, char *argv[]) {
    if (argc < 2) {
        die_with("nenhum comando especificado. comandos conhecidos: send, recv");
    }
    info("iniciando...\n");
    argv++;
    if (!strcmp("send", argv[0])) {
        /* die_with("send"); */
        handle_send();
    }
    if (!strcmp("recv", argv[0])) {
        /* die_with("recv"); */
        handle_recv();
    }
    die_with("comando '%s' não encontrado. comandos conhecidos: send, recv", argv[0]);
}

