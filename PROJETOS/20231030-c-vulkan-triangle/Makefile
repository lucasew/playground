GIT_ROOT := $(shell git rev-parse --show-toplevel)
VULKAN_INC := $(GIT_ROOT)/third-party/vulkan-headers/include
VULKAN_UTIL_INC := $(GIT_ROOT)/third-party/vulkan-utility-libraries/include

main: main.c frag.spv vert.spv
	gcc main.c -o a.out -I$(VULKAN_INC) -I$(VULKAN_UTIL_INC)/vulkan -lglfw -lvulkan -ldl -lpthread -lX11 -lXxf86vm -lXrandr -lXi -g

%.spv: shader.%
	glslc -o $@ $<


clean:
	rm a.out

all: a.out
