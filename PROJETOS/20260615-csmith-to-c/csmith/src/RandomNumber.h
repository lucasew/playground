// -*- mode: C++ -*-
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

#ifndef RANDOM_NUMBER_H
#define RANDOM_NUMBER_H

///////////////////////////////////////////////////////////////////////////////

#include "AbsRndNumGenerator.h"

#include <stdbool.h>

struct Filter;

/*
 * Common interface of all random number generators.
 * Works like a bridge to those generators. (C port)
 */
typedef struct RandomNumber {
  struct AbsRndNumGenerator *curr_generator_;
} RandomNumber;

void RandomNumber_CreateInstance(RNDNUM_GENERATOR rImpl, unsigned long seed);
RandomNumber *RandomNumber_GetInstance(void);
struct AbsRndNumGenerator *RandomNumber_GetRndNumGenerator(void);
RNDNUM_GENERATOR RandomNumber_SwitchRndNumGenerator(RNDNUM_GENERATOR rImpl);
void RandomNumber_doFinalization(void);

unsigned int RandomNumber_rnd_upto(RandomNumber *r, unsigned int n, const struct Filter *f, const char *where);
bool RandomNumber_rnd_flipcoin(RandomNumber *r, unsigned int p, const struct Filter *f, const char *where);

char *RandomNumber_RandomHexDigits(RandomNumber *r, int num);
char *RandomNumber_RandomDigits(RandomNumber *r, int num);

#endif // RANDOM_NUMBER_H
