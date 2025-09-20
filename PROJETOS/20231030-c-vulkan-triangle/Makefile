main: main.c frag.spv vert.spv
	gcc main.c -o a.out -lglfw -lvulkan -ldl -lpthread -lX11 -lXxf86vm -lXrandr -lXi -g

%.spv: shader.%
	glslc -o $@ $<


clean:
	rm a.out

all: a.out
