
main: ikkp_server.o error.o main.o
	gcc -Wall main.o ikkp_server.o error.o -o main

main.o: main.c ikkp_server.o error.o
	gcc -Wall -c main.c

ikkp_server.o: ikkp_server.c ikkp_server.h error.o
	gcc -Wall -c ikkp_server.c

error.o: error.c error.h
	gcc -Wall -c error.c

clean:
	rm *.o main
