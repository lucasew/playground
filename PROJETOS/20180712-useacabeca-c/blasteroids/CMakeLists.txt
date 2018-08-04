cmake_minimum_required(VERSION 3.10)

project(blasteroids C)

include_directories(include/)

add_library(blasteroids STATIC
    src/asteroid.c
    src/event.c
    src/spaceship.c)
target_link_libraries(blasteroids -lallegro -lallegro_primitives)

add_library(blasteroids-utils 
    src/utils/draw.c
    src/utils/log.c
    src/utils/signal.c
)
target_link_libraries(blasteroids-utils -lm)

add_executable(game main.c)
target_link_libraries(game blasteroids blasteroids-utils)
