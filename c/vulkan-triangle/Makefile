main: main.c frag.spv vert.spv
	gcc main.c -o main -lglfw -lvulkan -ldl -lpthread -lX11 -lXxf86vm -lXrandr -lXi

%.spv: shader.%
	glslc -o $@ $<


clean:
	rm main

all: main
