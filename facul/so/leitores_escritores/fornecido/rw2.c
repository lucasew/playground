/*
Sistema leitores-escritores, solução com priorização dos leitores.

Compilar com gcc -Wall rw2.c -o rw2 -lpthread

Carlos Maziero, DINF/UFPR 2020
*/

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <pthread.h>
#include <semaphore.h>

#define NUM_READERS 5
#define NUM_WRITERS 2

sem_t buffer ;		// semáforo de exclusão no buffer
int shared ;		// área compartilhada

int readers ;		// contador de leitores
sem_t mcount ;		// semáforo de exclusão no contador

// espera um tempo aleatório entre 0 e n segundos (float)
void espera (int n)
{
  sleep (random() % n) ;	// pausa entre 0 e n segundos (inteiro)
  usleep (random() % 1000000) ;	// pausa entre 0 e 1 segundo (float)
}

// corpo das tarefas leitoras
void *readerBody (void *id)
{
  long tid = (long) id ;

  while (1)
  {
    // entra na seção crítica
    sem_wait (&mcount) ;
    if (readers == 0)
      sem_wait (&buffer) ;
    readers++ ;
    sem_post (&mcount) ;

    // faz a leitura
    printf ("R%02ld: read %d (%d readers)\n", tid, shared, readers) ;
    espera (2) ;

    // sai da seção crítica
    sem_wait (&mcount) ;
    readers-- ;
    if (readers == 0)
      sem_post (&buffer) ;
    sem_post (&mcount) ;

    printf ("R%02ld: out\n", tid) ;
    espera (2) ;
  }
}

// corpo das tarefas escritoras
void *writerBody (void *id)
{
  long tid = (long) id ;

  while (1)
  {
    // entra na seção crítica
    sem_wait (&buffer) ;

    // certifica-se de que não há leitores ativos
    if (readers != 0)
    {
      printf ("ERRO\n") ;
      exit (1) ;
    }

    // faz a escrita
    shared = random () % 1000 ;
    printf ("\t\t\tW%02ld: wrote %d\n", tid, shared) ;
    espera (2) ;

    // sai da seção crítica
    sem_post (&buffer) ;

    printf ("\t\t\tW%02ld: out\n", tid) ;
    espera (2) ;
  }
}

int main (int argc, char *argv[])
{
  pthread_t reader [NUM_READERS] ;
  pthread_t writer [NUM_WRITERS] ;
  long i ;

  shared  = 0 ;
  readers = 0 ;

  // inicia semaforos
  sem_init (&buffer, 0, 1) ;
  sem_init (&mcount, 0, 1) ;

  // cria leitores
  for (i=0; i<NUM_READERS; i++)
    if (pthread_create (&reader[i], NULL, readerBody, (void *) i))
    {
      perror ("pthread_create") ;
      exit (1) ;
    }

  // cria escritores
  for (i=0; i<NUM_WRITERS; i++)
    if (pthread_create (&writer[i], NULL, writerBody, (void *) i))
    {
      perror ("pthread_create") ;
      exit (1) ;
    }

  // main encerra aqui
  pthread_exit (NULL) ;
}
