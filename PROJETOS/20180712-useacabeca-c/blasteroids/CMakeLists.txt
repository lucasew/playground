cmake_minimum_required(VERSION 3.10)

project(blasteroids C)

add_library(utils log_utils.c draw_utils.c signal_utils.c)
target_link_libraries(utils -lm)

add_executable(game main.c event.c spaceship.c asteroid.c)
target_link_libraries(game -lallegro_primitives -lallegro utils)
