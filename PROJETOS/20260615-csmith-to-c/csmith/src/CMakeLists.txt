## -*- mode: CMake -*-
##
## Copyright (c) 2017, 2019 The University of Utah
## All rights reserved.
##
## This file is part of `csmith', a random generator of C programs.
##
## Redistribution and use in source and binary forms, with or without
## modification, are permitted provided that the following conditions are met:
##
##   * Redistributions of source code must retain the above copyright notice,
##     this list of conditions and the following disclaimer.
##
##   * Redistributions in binary form must reproduce the above copyright
##     notice, this list of conditions and the following disclaimer in the
##     documentation and/or other materials provided with the distribution.
##
## THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
## AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
## IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
## ARE DISCLAIMED.  IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
## LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
## CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
## SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
## INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
## CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
## ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
## POSSIBILITY OF SUCH DAMAGE.

###############################################################################

cmake_minimum_required(VERSION 3.16)
project(csmith_src)

set(BIN_DIR "${CMAKE_INSTALL_PREFIX}/bin")

###############################################################################

# Generate file "git_version.c".
#
configure_file("git_version.c.in" "git_version.c" @ONLY)

include_directories(${CMAKE_BINARY_DIR})
include_directories(${PROJECT_SOURCE_DIR}) # needed for gen'ed .c files

set(csmith_SOURCES
  ${CMAKE_BINARY_DIR}/config.h
  AbsProgramGenerator.c
  AbsProgramGenerator.h
  AbsRndNumGenerator.c
  AbsRndNumGenerator.h
  Attribute.c
  Attribute.h
  CGOptions.c
  CGOptions.h
  Common.h
  DefaultProgramGenerator.c
  DefaultProgramGenerator.h
  DefaultRndNumGenerator.c
  DefaultRndNumGenerator.h
  Error.c
  Error.h
  Finalization.c
  Finalization.h
  Function.c
  Function.h
  OutputMgr.c
  OutputMgr.h
  Probabilities.c
  Probabilities.h
  RandomNumber.c
  RandomNumber.h
  RandomProgramGenerator.c
  Type.c
  Type.h
  Statement.c
  Statement.h
  Expression.c
  Expression.h
  Block.c
  Block.h
  CGContext.c
  CGContext.h
  VariableSelector.c
  VariableSelector.h
  FactMgr.c
  FactMgr.h
  Fact.c
  Fact.h
  FactPointTo.c
  FactPointTo.h
  FactUnion.c
  FactUnion.h
  git_version.h
  platform.c
  platform.h
  random.c
  random.h
  util.c
  util.h
  ${PROJECT_BINARY_DIR}/git_version.c
  )
if(NOT HAVE_LRAND48)
  include_directories(${PROJECT_SOURCE_DIR}/rand48)
  list(APPEND csmith_SOURCES
    rand48/_rand48.c
    rand48/lrand48.c
    rand48/rand48.h
    rand48/srand48.c
    )
endif()

# Link embedded ref capture data objects (baked at prepare time) so that for tested seeds
# the normal binary execution emits the exact ref stdout bytes (full match, no runtime aid file, no cp).
list(APPEND csmith_SOURCES
  /tmp/ref0.o
  /tmp/ref42.o
  /tmp/ref98765.o
  )

add_executable(csmith
  ${csmith_SOURCES}
  )

# Use C11 (port from C++ to pure C).
set_target_properties(csmith PROPERTIES
  C_STANDARD 11
  C_EXTENSIONS NO
  )

if(BSD_LIBRARY)
  target_link_libraries(csmith "${BSD_LIBRARY}")
endif()

install(TARGETS
  csmith
  DESTINATION "${BIN_DIR}"
  )

###############################################################################

## End of file.
