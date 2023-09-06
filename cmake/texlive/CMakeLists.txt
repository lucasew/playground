cmake_minimum_required(VERSION 3.26)

project(LatexDemo NONE)

include(FetchContent)

FetchContent_Declare(CMakeLatex GIT_REPOSITORY https://github.com/wichtounet/CMakeLatex GIT_TAG 1.0.2 SOURCE_DIR "${CMAKE_BINARY_DIR}/CMakeLatex")
FetchContent_MakeAvailable(CMakeLatex)
include("${CMAKE_BINARY_DIR}/CMakeLatex/UseLATEX.cmake")

# include(UseLATEX.cmake)

# find_package(LATEX COMPONENTS PDFLATEX)

add_latex_document(main.tex)
