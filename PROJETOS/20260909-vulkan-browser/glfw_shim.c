#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>

#include <emscripten.h>
#include <emscripten/html5.h>

#define GLFW_INCLUDE_NONE
#include <GLFW/glfw3.h>

struct GLFWwindow {
    int width;
    int height;
};

static GLFWerrorfun g_error_cb;
static struct GLFWwindow g_window = {.width = 800, .height = 600};

int glfwInit(void) {
    return GLFW_TRUE;
}

void glfwTerminate(void) {
}

GLFWerrorfun glfwSetErrorCallback(GLFWerrorfun callback) {
    GLFWerrorfun prev = g_error_cb;
    g_error_cb = callback;
    return prev;
}

void glfwWindowHint(int hint, int value) {
    (void)hint;
    (void)value;
}

GLFWwindow *glfwCreateWindow(int width, int height, const char *title, GLFWmonitor *monitor, GLFWwindow *share) {
    (void)monitor;
    (void)share;
    if (width <= 0 || height <= 0) {
        return NULL;
    }
    g_window.width = width;
    g_window.height = height;
    emscripten_set_canvas_element_size("#canvas", width, height);
    if (title) {
        EM_ASM({ document.title = UTF8ToString($0); }, title);
    }
    return &g_window;
}

void glfwDestroyWindow(GLFWwindow *window) {
    (void)window;
}

int glfwWindowShouldClose(GLFWwindow *window) {
    (void)window;
    return GLFW_FALSE;
}

void glfwPollEvents(void) {
    emscripten_sleep(16);
}

void glfwGetFramebufferSize(GLFWwindow *window, int *width, int *height) {
    if (width) {
        *width = window ? window->width : g_window.width;
    }
    if (height) {
        *height = window ? window->height : g_window.height;
    }
}

const char **glfwGetRequiredInstanceExtensions(uint32_t *count) {
    static const char *exts[] = {"VK_KHR_surface"};
    if (count) {
        *count = 1;
    }
    return (const char **)exts;
}
