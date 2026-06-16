// -*- mode: C -*-
//
// Copyright (c) 2007, 2008, 2009, 2010, 2011 The University of Utah
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

#ifndef DEFAULT_RNDNUM_GENERATOR_H
#define DEFAULT_RNDNUM_GENERATOR_H

#include "AbsRndNumGenerator.h"
#include "Common.h"

struct Sequence;
struct Filter;

// C port: struct holding the state for default rnd generator.
typedef struct DefaultRndNumGeneratorData {
  unsigned INT64 rand_depth_;
  char *trace_string_;   // owned dynamic
  struct Sequence *seq_;
} DefaultRndNumGeneratorData;

AbsRndNumGenerator *DefaultRndNumGenerator_make_rndnum_generator(const unsigned long seed);

void DefaultRndNumGenerator_set_rand_depth(AbsRndNumGenerator *g, unsigned INT64 depth);

void DefaultRndNumGenerator_destroy(AbsRndNumGenerator *g);

#endif // DEFAULT_RNDNUM_GENERATOR_H
