# vulkan-browser

The 2023 `main.c` from `20231030-c-vulkan-triangle`, unchanged. It still talks
Vulkan + GLFW. A small ICD (`vulkan_webgpu.c`) and GLFW shim (`glfw_shim.c`)
turn those calls into WebGPU so Emscripten can run the same file in a browser.

## Build

Emscripten 4.0.9 and Binaryen 123 come from `mise.toml` (conda + HTTP backends).
Vulkan headers and `vk_enum_string_helper.h` come from the playground-root
submodules `third-party/vulkan-headers` and
`third-party/vulkan-utility-libraries` (`git submodule update --init`).

```bash
mise run build
```

That writes `dist/index.html`, `dist/index.js`, and `dist/index.wasm`.

## Run

WebGPU needs a secure origin. `localhost` counts.

```bash
mise run serve
```

`goftpd --spa` serves `dist/` on http://localhost:8080/. Open that in a browser
with WebGPU (Chrome, Edge, or Firefox 141+).

You should see a black 800x600 canvas and a triangle: red tip up, green bottom-right, blue bottom-left.

## What stays, what is new

`main.c`, `shader.vert`, and `shader.frag` are symlinks to
`../20231030-c-vulkan-triangle`. Do not edit `main.c` to "port" it.

The shim implements the Vulkan entry points that file actually calls. Shader
modules still load `vert.spv` / `frag.spv` through `mmap` as before. The SPIR-V
is only used to tell vertex from fragment; the draw uses WGSL that matches the
GLSL (Y flipped, because Vulkan NDC is Y-down and WebGPU is Y-up).

The `while (!glfwWindowShouldClose)` loop is kept. `glfwPollEvents` yields with
Emscripten ASYNCIFY so the tab stays responsive.
