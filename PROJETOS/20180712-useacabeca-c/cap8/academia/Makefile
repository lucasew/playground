INCLUDE=-I./includes
LIB=-L./libs
VARIANT=UK

elliptical: elliptical.o libs/libhfcal.a
	gcc elliptical.o ${INCLUDE} ${LIB} -lhfcal -o elliptical

hfcal_${VARIANT}.o: hfcal_${VARIANT}.c
	$(info Compilando com a variante ${VARIANT})
	gcc -c -fPIC hfcal_${VARIANT}.c ${INCLUDE} -o hfcal.o

elliptical.o: elliptical.c
	gcc -c elliptical.c ${INCLUDE}

libs/:
	mkdir libs

libs/libhfcal.a: hfcal_${VARIANT}.o libs/
	ar -rcs ./libs/libhfcal.a hfcal.o

clean:
	rm -rf libs/ *.o elliptical
