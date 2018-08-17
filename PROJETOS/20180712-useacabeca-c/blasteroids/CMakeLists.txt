cmake_minimum_required(VERSION 3.10)

project(blasteroids LANGUAGES C)

include_directories(include/)

add_library(blasteroids STATIC
    src/asteroid.c
    src/bullet.c
    src/collision.c
    src/context.c
    src/event.c
    src/spaceship.c
    src/pos_fixer.c
    src/text.c
)
target_link_libraries(blasteroids -lallegro -lallegro_ttf -lallegro_font -lallegro_primitives)

add_library(blasteroids-utils 
    src/utils/draw.c
    src/utils/signal.c
)
target_link_libraries(blasteroids-utils -lm)

add_executable(game main.c)
target_link_libraries(game blasteroids blasteroids-utils)
