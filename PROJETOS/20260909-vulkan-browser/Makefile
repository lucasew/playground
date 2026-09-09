EMCC ?= emcc
PYTHON ?= python3
OUT := dist/index.html
VULKAN_INC := third_party/vulkan-headers/include
BINARYEN_ROOT ?= $(shell mise where http:binaryen)

SRCS := main.c vulkan_webgpu.c glfw_shim.c

EMFLAGS := \
	-I$(VULKAN_INC) \
	-sUSE_WEBGPU=1 \
	-sASYNCIFY=1 \
	-sASYNCIFY_STACK_SIZE=65536 \
	-sALLOW_MEMORY_GROWTH=1 \
	-sASSERTIONS=1 \
	-sNO_EXIT_RUNTIME=1 \
	--shell-file shell.html \
	--embed-file vert.spv \
	--embed-file frag.spv

.PHONY: all clean serve

all: $(OUT)

vert.spv: tools/write_spv.py
	$(PYTHON) tools/write_spv.py vert $@

frag.spv: tools/write_spv.py
	$(PYTHON) tools/write_spv.py frag $@

$(OUT): $(SRCS) vk_enum_string_helper.h vert.spv frag.spv shell.html
	mkdir -p dist
	mise install
	EM_BINARYEN_ROOT="$$(mise where http:binaryen)" $(EMCC) $(SRCS) -o $(OUT) $(EMFLAGS) -O0 -g

serve: $(OUT)
	$(PYTHON) -m http.server --directory dist 8080

clean:
	rm -rf dist vert.spv frag.spv
