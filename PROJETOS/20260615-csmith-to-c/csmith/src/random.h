// -*- mode: C++ -*-
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

#ifndef RANDOM_H
#define RANDOM_H

///////////////////////////////////////////////////////////////////////////////



struct Filter;

// Old stuff.
char *RandomHexDigits(int num);
char *RandomDigits(int num);

// New stuff.
unsigned int rnd_upto(const unsigned int n, const struct Filter *f,
                      const char *where);
int rnd_flipcoin(const unsigned int p, const struct Filter *f,
                  const char *where);
// return pure random numbers even if csmith is in other mode, e.g., exhaustive
// mode
char *PureRandomHexDigits(int num);
char *PureRandomDigits(int num);
unsigned int pure_rnd_upto(const unsigned int n, const struct Filter *f,
                           const char *where);
int pure_rnd_flipcoin(const unsigned int p, const struct Filter *f,
                       const char *where);

char *get_prefixed_name(const char *name);
char *trace_depth(void);
void get_sequence(char **sequence);
#if 0
// Deprecated
unsigned int*   rnd_shuffle(unsigned int n);
#endif

///////////////////////////////////////////////////////////////////////////////

#endif // RANDOM_H

// Local Variables:
// c-basic-offset: 4
// tab-width: 4
// End:

// End of file.
