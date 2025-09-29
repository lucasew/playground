#include <stdio.h>
#include <stdlib.h>
#include <mqueue.h>
#include <string.h>
#include <stdarg.h>
#include <errno.h>
#include <signal.h>
#include <semaphore.h>
#include <unistd.h>
#include <pthread.h>

// tá mei zoado ainda, se der tempo eu arrumo

#define die_with(msg, args...) fprintf(stderr, "\033[0;31merro\033[0m: "); fprintf(stderr, msg, ## args); exit(1);
#define die_with_errno() die_with("%s", strerror(errno))
#define info(msg, args...) fprintf(stderr, "\033[0;32minfo\033[0m: "); fprintf(stderr, msg, ## args);

#define NUM_READERS 3
#define NUM_WRITERS 1

#define VECMAX 51

int incr(int indice) {
    return (indice + 1)%VECMAX;
}
int shuffle() {
    return random() % 100;
}

typedef struct fila_t {
    int vec[VECMAX];
    short inicio;
    short final;
} fila_t;

fila_t *l = NULL;

fila_t *fila_new() {
    fila_t *f = malloc(sizeof(fila_t));
    f->inicio = 0;
    f->final = 0;
    return f;
}

int fila_tamanho(fila_t *f) {
    return (VECMAX - f->inicio + f->final)%VECMAX;
}

int fila_isvazio(fila_t *f) {
    return fila_tamanho(f) == 0;
}

int fila_ischeio(fila_t *f) {
    return fila_tamanho(f) == VECMAX - 1;
}

void fila_insert(fila_t *f, int v) {
    if (fila_ischeio(f)) {
        printf("E: Fila cheia\n");
        exit(1);
    }
    f->vec[f->final] = v;
    f->final = incr(f->final);
}

int fila_next(fila_t *f) {
    int v;
    if(fila_isvazio(f)) {
        printf("E: Fila vazia\n");
        exit(1);
    }
    v = f->vec[f->inicio];
    f->inicio = incr(f->inicio);
    return v;
}

void fila_destroy(fila_t *f) {
    free(f);
}


int numReaders = 0;
int writeMode = 0;
sem_t sem_read;
sem_t sem_write;

void rw_lock() {
    printf("rwlock writeMode: %i numReaders: %i\n", writeMode, numReaders);
    sem_wait(&sem_read);
    sem_wait(&sem_write);
    writeMode = 1;
    sem_post(&sem_read);
}

void r_lock() {
    printf("rlock writeMode: %i numReaders: %i\n", writeMode, numReaders);
    sem_wait(&sem_read);
    int noReaders = numReaders == 0;
    if (noReaders) {
        sem_post(&sem_read);
        sem_wait(&sem_write);
        sem_wait(&sem_read);
        numReaders++;
        writeMode = 0;
    } else {
        numReaders++;
    }
    sem_post(&sem_read);
}

void unlock() {
    printf("unlock writeMode: %i numReaders: %i\n", writeMode, numReaders);
    sem_wait(&sem_read);
    if (writeMode == 0) {
        if (numReaders == 0) {
            sem_post(&sem_write);
        } else {
            numReaders--;
        }
    } else {
        sem_post(&sem_write);
        writeMode = 0;
    }
    sem_post(&sem_read);
}

void prio_reader() {}
void prio_writer() {}
void prio_same() {}

int keepRunning = 1;

pthread_t readers[NUM_READERS];
pthread_t writers[NUM_WRITERS];


void* read_handler(void*v) {
    while (keepRunning) {
        r_lock();
        // TODO: printar lista
        unlock();
        sleep(1);
    }
    return NULL;
}

void* write_handler(void*v) {
    while (keepRunning) {
        rw_lock();
        fila_next(l);
        int shuffled = shuffle();
        info("shuffled: %i\n", shuffled);
        fila_insert(l, shuffled);
        unlock();
        sleep(1);
    }
    return NULL;
}

int main(int argc, char *argv[]) {
    srand(time(NULL));
    setvbuf (stdout, 0, _IONBF, 0) ;

    l = fila_new();
    fila_insert(l, 1);
    fila_insert(l, 2);
    fila_insert(l, 3);
    info("iniciando...\n");
    if (argc < 2) {
        die_with("nenhum comando especificado. comandos conhecidos: r, w, rw");
    }
    argv++;
    if (!strcmp("r", argv[0])) {
        prio_reader();
    } else if (!strcmp("w", argv[0])) {
        prio_writer();
    } else if (!strcmp("rw", argv[0])) {
        prio_same();
    } else {
        die_with("comando '%s' não encontrado. comandos conhecidos: r, w, rw", argv[0]);
    }
    sem_init(&sem_read, 0, 1);
    sem_init(&sem_write, 0, 1);
    for (int i = 0; i < NUM_READERS; i++) {
        if(pthread_create(&readers[i], NULL, read_handler, NULL)) {
            die_with_errno();
        }
    }
    for (int i = 0; i < NUM_WRITERS; i++) {
        if(pthread_create(&writers[i], NULL, write_handler, NULL)) {
            die_with_errno();
        }
    }
    /* lista_destroy(l); */
    pthread_exit(NULL);
}

