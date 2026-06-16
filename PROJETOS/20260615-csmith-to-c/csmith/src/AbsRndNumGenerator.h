// -*- mode: C -*-
//
// Copyright (c) 2007, 2008, 2009, 2010, 2011, 2015 The University of Utah
// All rights reserved.
//
// This file is part of `csmith', a random generator of C programs.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
//   * Redistributions of source code must retain the above copyright notice,
//     this list of conditions and the following disclaimer.
//
//   * Redistributions in binary form must reproduce the above copyright
//     notice, this list of conditions and the following disclaimer in the
//     documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED.  IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

#ifndef ABS_RNDNUM_GENERATOR
#define ABS_RNDNUM_GENERATOR

#include <stdbool.h>

struct Filter;

typedef enum {
  rDefaultRndNumGenerator = 0,
  rDFSRndNumGenerator,
} RNDNUM_GENERATOR;

#define MAX_RNDNUM_GENERATOR (rDFSRndNumGenerator + 1)

// C port: use tagged structs + dispatch funcs instead of base class.
typedef struct AbsRndNumGenerator {
  RNDNUM_GENERATOR kind_tag;
  // data per impl filled in concrete
  void *impl_data; // e.g. points to DefaultRndNumGeneratorData or similar
} AbsRndNumGenerator;

AbsRndNumGenerator *AbsRndNumGenerator_make_rndnum_generator(RNDNUM_GENERATOR impl,
                                                             const unsigned long seed);

void AbsRndNumGenerator_seedrand(const unsigned long seed);

const char *AbsRndNumGenerator_get_hex1(void);

const char *AbsRndNumGenerator_get_dec1(void);

unsigned int AbsRndNumGenerator_count(void);

char *AbsRndNumGenerator_get_prefixed_name(AbsRndNumGenerator *g, const char *name); // caller frees? or use static buf for simplicity

char *AbsRndNumGenerator_trace_depth(AbsRndNumGenerator *g);

void AbsRndNumGenerator_get_sequence(AbsRndNumGenerator *g, char **sequence /* out, caller manages */);

unsigned int AbsRndNumGenerator_rnd_upto(AbsRndNumGenerator *g, const unsigned int n, const struct Filter *f,
                                         const char *where);

bool AbsRndNumGenerator_rnd_flipcoin(AbsRndNumGenerator *g, const unsigned int p, const struct Filter *f,
                                     const char *where);

char *AbsRndNumGenerator_RandomHexDigits(AbsRndNumGenerator *g, int num);

char *AbsRndNumGenerator_RandomDigits(AbsRndNumGenerator *g, int num);

RNDNUM_GENERATOR AbsRndNumGenerator_kind(AbsRndNumGenerator *g);

void AbsRndNumGenerator_destroy(AbsRndNumGenerator *g);

unsigned long AbsRndNumGenerator_genrand(void); // the protected one, exposed for impls

// internal hex/dec digits
extern const char *abs_rnd_hex1;
extern const char *abs_rnd_dec1;

#endif // ABS_RNDNUM_GENERATOR
