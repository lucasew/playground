cmake_minimum_required(VERSION 3.10)

project(blasteroids LANGUAGES C)

include_directories(include/)

add_executable(game
    main.c
    src/asteroid.c
    src/bullet.c
    src/collision.c
    src/context.c
    src/event.c
    src/spaceship.c
    src/pos_fixer.c
    src/text.c
    src/utils/draw.c
    src/utils/signal.c
)

target_link_libraries(game -lallegro -lallegro_ttf -lallegro_font -lallegro_primitives -lm)

