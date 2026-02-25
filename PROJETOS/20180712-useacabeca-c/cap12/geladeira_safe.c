#include <pthread.h>
#include <stdio.h>

int beers = 2000000;
pthread_mutex_t a_lock = PTHREAD_MUTEX_INITIALIZER;

void* drink_lots(void *a) {
    int i;
    for (i = 0; i < 100000; i++) {
        pthread_mutex_lock(&a_lock);
        beers = beers - 1;
        pthread_mutex_unlock(&a_lock);
    }
    printf("beers = %i\n", beers);
    return NULL;
}

int main() {
    pthread_t threads[20]; // 20 escravos :p
    int t;
    printf("%i bottles of beer on the wall\n %i bottles of beer\n", beers, beers);
    for (t = 0; t < 20; t++) {
        pthread_create(&threads[t], NULL, drink_lots, NULL);
    }
    void* result;
    for (t = 0; t < 20; t++) {
        pthread_join(threads[t], &result);
    }
    printf("There are now %i bottles of beer on the wall!\n", beers);
    return 0;
}
