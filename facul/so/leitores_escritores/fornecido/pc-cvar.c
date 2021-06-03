/*
Sistema produtor-consumidor usando variáveis de condição.

ATENÇÃO: este exemplo está mais simples que aquele que aparece no vídeo.

Compilar com gcc -Wall pc-cvar.c -o pc-cvar lpthread

Carlos Maziero, DINF/UFPR 2020
*/

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <pthread.h>

#define NUM_PROD 5
#define NUM_CONS 3
#define VAGAS 6

pthread_mutex_t mutex ;			// mutex para acesso ao buffer
pthread_cond_t itens_cv ;		// vc para controle de itens
pthread_cond_t vagas_cv ;		// vc para controle de vagas

int num_itens, num_vagas ; 		// contadores de itens e vagas

// espera um tempo aleatório entre 0 e n segundos (float)
void espera (int n)
{
  sleep (random() % n) ;	// pausa entre 0 e n segundos (inteiro)
  usleep (random() % 1000000) ;	// pausa entre 0 e 1 segundo (float)
}

// corpo de produtor
void *prodBody (void *id)
{
  long tid = (long) id ;

  while (1)
  {
    // requisita acesso exclusivo ao buffer
    pthread_mutex_lock (&mutex);

    // se não houver vaga, libera o buffer e espera
    while (num_vagas == 0)
      pthread_cond_wait (&vagas_cv, &mutex);

    // coloca um item no buffer
    num_vagas-- ;
    num_itens++ ;
    printf ("P%02ld: put item (%d items, %d places)\n",
            tid, num_itens, num_vagas) ;

    // sinaliza um novo item
    pthread_cond_signal (&itens_cv);

    // libera o buffer
    pthread_mutex_unlock (&mutex) ;

    espera (2) ;
  }
}

// corpo de cosumidor
void *consBody (void *id)
{
  long tid = (long) id ;

  while (1)
  {
    // requisita acesso exclusivo ao buffer
    pthread_mutex_lock (&mutex) ;

    // se não houver item, libera o buffer e espera
    while (num_itens == 0)
      pthread_cond_wait (&itens_cv, &mutex) ;

    // retira um item do buffer
    num_itens-- ;
    num_vagas++ ;
    printf ("\t\t\t\t\tC%02ld: got item (%d items, %d places)\n",
            tid, num_itens, num_vagas) ;

    // sinaliza uma nova vaga
    pthread_cond_signal (&vagas_cv) ;

    // libera o buffer
    pthread_mutex_unlock (&mutex) ;

    espera (2) ;
  }
}

// programa principal
int main (int argc, char *argv[])
{
  pthread_t produtor   [NUM_PROD] ;
  pthread_t consumidor [NUM_CONS] ;
  long i ;

  // inicia contadores
  num_itens = 0 ;
  num_vagas = VAGAS ;

  // inicia variaveis de condição e mutexes
  pthread_mutex_init (&mutex, NULL) ;
  pthread_cond_init (&itens_cv, NULL) ;
  pthread_cond_init (&vagas_cv, NULL) ;

  // cria produtores
  for (i=0; i<NUM_PROD; i++)
    if (pthread_create (&produtor[i], NULL, prodBody, (void *) i))
    {
      perror ("pthread_create") ;
      exit (1) ;
    }

  // cria consumidores
  for (i=0; i<NUM_CONS; i++)
    if (pthread_create (&consumidor[i], NULL, consBody, (void *) i))
    {
      perror ("pthread_create") ;
      exit (1) ;
    }

  // main encerra aqui
  pthread_exit (NULL) ;
}