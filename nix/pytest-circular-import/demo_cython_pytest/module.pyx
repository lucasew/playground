#cython: language_level=3

cimport cython

def native_sum(x, y):
    return cy_sum(x, y)

cdef int cy_sum(int x, int y):
    return x + y
