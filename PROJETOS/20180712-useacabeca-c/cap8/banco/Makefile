EXEFILE=banco


${EXEFILE}: libhfsecurity main.c
	gcc main.c -I. -L. libhfsecurity.a -o ${EXEFILE}

libhfsecurity: checksum.o encrypt.o
	ar -rcs libhfsecurity.a encrypt.o checksum.o

checksum.o: checksum.c
	gcc -Wall -c checksum.c

encrypt.o: encrypt.c
	gcc -Wall -c encrypt.c

clean:
	rm -f *.o *.a {EXEFILE}
