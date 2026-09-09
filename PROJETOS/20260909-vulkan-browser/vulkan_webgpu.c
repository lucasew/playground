#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

#include <emscripten.h>
#include <webgpu/webgpu.h>
#include <emscripten/html5_webgpu.h>

#define GLFW_INCLUDE_VULKAN
#include <GLFW/glfw3.h>

#define ND_HANDLE(p) ((uint64_t)(uintptr_t)(p))
#define ND_PTR(type, h) ((type *)(uintptr_t)(h))

#define MAX_SWAP_IMAGES 8

static const char *kVertexWGSL =
    "struct VSOut {\n"
    "  @builtin(position) position: vec4<f32>,\n"
    "  @location(0) fragColor: vec3<f32>,\n"
    "}\n"
    "@vertex\n"
    "fn main(@builtin(vertex_index) idx: u32) -> VSOut {\n"
    "  var positions = array<vec2<f32>, 3>(\n"
    "    vec2<f32>(0.0, 0.5),\n"
    "    vec2<f32>(0.5, -0.5),\n"
    "    vec2<f32>(-0.5, -0.5)\n"
    "  );\n"
    "  var colors = array<vec3<f32>, 3>(\n"
    "    vec3<f32>(1.0, 0.0, 0.0),\n"
    "    vec3<f32>(0.0, 1.0, 0.0),\n"
    "    vec3<f32>(0.0, 0.0, 1.0)\n"
    "  );\n"
    "  var out: VSOut;\n"
    "  out.position = vec4<f32>(positions[idx], 0.0, 1.0);\n"
    "  out.fragColor = colors[idx];\n"
    "  return out;\n"
    "}\n";

static const char *kFragmentWGSL =
    "@fragment\n"
    "fn main(@location(0) fragColor: vec3<f32>) -> @location(0) vec4<f32> {\n"
    "  return vec4<f32>(fragColor, 1.0);\n"
    "}\n";

struct VkInstance_T {
    WGPUInstance wgpu;
    struct VkPhysicalDevice_T *phys;
};

struct VkPhysicalDevice_T {
    struct VkInstance_T *instance;
};

struct VkDevice_T {
    WGPUDevice wgpu;
    WGPUQueue queue;
    struct VkPhysicalDevice_T *phys;
    struct VkQueue_T *vk_queue;
};

struct VkQueue_T {
    WGPUQueue wgpu;
};

struct VkSurfaceKHR_T {
    WGPUSurface wgpu;
    uint32_t width;
    uint32_t height;
};

struct VkImageView_T;

struct VkImage_T {
    WGPUTexture texture;
    struct VkImageView_T *view;
};

struct VkImageView_T {
    struct VkImage_T *image;
    WGPUTextureView wgpu;
};

struct VkSwapchainKHR_T {
    struct VkDevice_T *device;
    struct VkSurfaceKHR_T *surface;
    uint32_t image_count;
    struct VkImage_T *images[MAX_SWAP_IMAGES];
    WGPUTextureFormat wgpu_format;
    VkFormat vk_format;
    uint32_t width;
    uint32_t height;
    WGPUTexture current_texture;
};

struct VkShaderModule_T {
    WGPUShaderModule wgpu;
    int is_vertex;
};

struct VkPipelineLayout_T {
    int dummy;
};

struct VkRenderPass_T {
    VkFormat format;
};

struct VkPipeline_T {
    WGPURenderPipeline wgpu;
};

struct VkFramebuffer_T {
    struct VkImageView_T *color;
    uint32_t width;
    uint32_t height;
};

struct VkCommandPool_T {
    struct VkDevice_T *device;
};

struct VkCommandBuffer_T {
    struct VkDevice_T *device;
    struct VkFramebuffer_T *framebuffer;
    struct VkPipeline_T *pipeline;
    VkViewport viewport;
    VkRect2D scissor;
    int has_viewport;
    int has_scissor;
    uint32_t draw_vertices;
};

struct VkSemaphore_T {
    int dummy;
};

struct VkFence_T {
    int signaled;
};

static int g_shader_seq;
static struct VkSwapchainKHR_T *g_active_swapchain;

EM_ASYNC_JS(int, js_setup_webgpu, (), {
    if (!navigator.gpu) {
        return 1;
    }
    const adapter = await navigator.gpu.requestAdapter();
    if (!adapter) {
        return 2;
    }
    const device = await adapter.requestDevice();
    if (!device) {
        return 3;
    }
    Module.preinitializedWebGPUDevice = device;
    return 0;
});

static void on_wgpu_error(WGPUErrorType type, const char *message, void *userdata) {
    (void)userdata;
    fprintf(stderr, "webgpu error %u: %s\n", (unsigned)type, message ? message : "(null)");
}

static WGPUTextureFormat canvas_format(void) {
    /* Browsers reject bgra8unorm-srgb as a canvas context format. */
    return WGPUTextureFormat_BGRA8Unorm;
}

static int spirv_execution_model(const uint32_t *code, size_t bytes) {
    if (!code || bytes < 20 || code[0] != 0x07230203u) {
        return -1;
    }
    size_t nwords = bytes / 4;
    size_t i = 5;
    while (i < nwords) {
        uint32_t word = code[i];
        uint32_t opcode = word & 0xffffu;
        uint32_t count = word >> 16;
        if (count == 0 || i + count > nwords) {
            break;
        }
        if (opcode == 15 && count >= 2) {
            return (int)code[i + 1];
        }
        i += count;
    }
    return -1;
}

static WGPUShaderModule create_wgsl_module(WGPUDevice device, const char *code) {
    WGPUShaderModuleWGSLDescriptor wgsl = {
        .chain = {.sType = WGPUSType_ShaderModuleWGSLDescriptor},
        .code = code,
    };
    WGPUShaderModuleDescriptor desc = {
        .nextInChain = &wgsl.chain,
    };
    return wgpuDeviceCreateShaderModule(device, &desc);
}

static void refresh_image_view(struct VkImage_T *image) {
    if (!image || !image->view || !image->texture) {
        return;
    }
    if (image->view->wgpu) {
        wgpuTextureViewRelease(image->view->wgpu);
        image->view->wgpu = NULL;
    }
    image->view->wgpu = wgpuTextureCreateView(image->texture, NULL);
}

VkResult vkEnumerateInstanceExtensionProperties(const char *pLayerName, uint32_t *pPropertyCount, VkExtensionProperties *pProperties) {
    (void)pLayerName;
    static const VkExtensionProperties exts[] = {
        {.extensionName = "VK_KHR_surface", .specVersion = 25},
        {.extensionName = "VK_KHR_swapchain", .specVersion = 70},
    };
    if (!pPropertyCount) {
        return VK_ERROR_INITIALIZATION_FAILED;
    }
    if (!pProperties) {
        *pPropertyCount = 2;
        return VK_SUCCESS;
    }
    uint32_t n = *pPropertyCount < 2 ? *pPropertyCount : 2;
    memcpy(pProperties, exts, n * sizeof(exts[0]));
    *pPropertyCount = n;
    return VK_SUCCESS;
}

VkResult vkEnumerateInstanceLayerProperties(uint32_t *pPropertyCount, VkLayerProperties *pProperties) {
    if (!pPropertyCount) {
        return VK_ERROR_INITIALIZATION_FAILED;
    }
    if (!pProperties) {
        *pPropertyCount = 0;
        return VK_SUCCESS;
    }
    *pPropertyCount = 0;
    return VK_SUCCESS;
}

VkResult vkCreateInstance(const VkInstanceCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkInstance *pInstance) {
    (void)pCreateInfo;
    (void)pAllocator;
    struct VkInstance_T *inst = calloc(1, sizeof(*inst));
    struct VkPhysicalDevice_T *phys = calloc(1, sizeof(*phys));
    if (!inst || !phys) {
        free(inst);
        free(phys);
        return VK_ERROR_OUT_OF_HOST_MEMORY;
    }
    inst->wgpu = wgpuCreateInstance(NULL);
    if (!inst->wgpu) {
        free(inst);
        free(phys);
        return VK_ERROR_INITIALIZATION_FAILED;
    }
    if (js_setup_webgpu() != 0) {
        fprintf(stderr, "webgpu: navigator.gpu / requestAdapter failed\n");
        wgpuInstanceRelease(inst->wgpu);
        free(inst);
        free(phys);
        return VK_ERROR_INITIALIZATION_FAILED;
    }
    phys->instance = inst;
    inst->phys = phys;
    *pInstance = inst;
    return VK_SUCCESS;
}

void vkDestroyInstance(VkInstance instance, const VkAllocationCallbacks *pAllocator) {
    (void)pAllocator;
    if (!instance) {
        return;
    }
    free(instance->phys);
    if (instance->wgpu) {
        wgpuInstanceRelease(instance->wgpu);
    }
    free(instance);
}

PFN_vkVoidFunction vkGetInstanceProcAddr(VkInstance instance, const char *pName) {
    (void)instance;
    (void)pName;
    return NULL;
}

VkResult vkEnumeratePhysicalDevices(VkInstance instance, uint32_t *pPhysicalDeviceCount, VkPhysicalDevice *pPhysicalDevices) {
    if (!instance || !pPhysicalDeviceCount) {
        return VK_ERROR_INITIALIZATION_FAILED;
    }
    if (!pPhysicalDevices) {
        *pPhysicalDeviceCount = 1;
        return VK_SUCCESS;
    }
    if (*pPhysicalDeviceCount < 1) {
        return VK_INCOMPLETE;
    }
    pPhysicalDevices[0] = instance->phys;
    *pPhysicalDeviceCount = 1;
    return VK_SUCCESS;
}

void vkGetPhysicalDeviceProperties(VkPhysicalDevice physicalDevice, VkPhysicalDeviceProperties *pProperties) {
    (void)physicalDevice;
    memset(pProperties, 0, sizeof(*pProperties));
    pProperties->apiVersion = VK_API_VERSION_1_0;
    pProperties->driverVersion = 1;
    pProperties->vendorID = 0x1;
    pProperties->deviceID = 0x1;
    pProperties->deviceType = VK_PHYSICAL_DEVICE_TYPE_DISCRETE_GPU;
    strncpy(pProperties->deviceName, "WebGPU", VK_MAX_PHYSICAL_DEVICE_NAME_SIZE - 1);
    pProperties->limits.maxImageDimension2D = 8192;
}

void vkGetPhysicalDeviceFeatures(VkPhysicalDevice physicalDevice, VkPhysicalDeviceFeatures *pFeatures) {
    (void)physicalDevice;
    memset(pFeatures, 0, sizeof(*pFeatures));
    pFeatures->geometryShader = VK_TRUE;
}

void vkGetPhysicalDeviceQueueFamilyProperties(VkPhysicalDevice physicalDevice, uint32_t *pQueueFamilyPropertyCount, VkQueueFamilyProperties *pQueueFamilyProperties) {
    (void)physicalDevice;
    if (!pQueueFamilyProperties) {
        *pQueueFamilyPropertyCount = 1;
        return;
    }
    if (*pQueueFamilyPropertyCount < 1) {
        return;
    }
    memset(pQueueFamilyProperties, 0, sizeof(*pQueueFamilyProperties));
    pQueueFamilyProperties[0].queueFlags = VK_QUEUE_GRAPHICS_BIT | VK_QUEUE_COMPUTE_BIT | VK_QUEUE_TRANSFER_BIT;
    pQueueFamilyProperties[0].queueCount = 1;
    *pQueueFamilyPropertyCount = 1;
}

VkResult vkGetPhysicalDeviceSurfaceSupportKHR(VkPhysicalDevice physicalDevice, uint32_t queueFamilyIndex, VkSurfaceKHR surface, VkBool32 *pSupported) {
    (void)physicalDevice;
    (void)queueFamilyIndex;
    (void)surface;
    if (pSupported) {
        *pSupported = VK_TRUE;
    }
    return VK_SUCCESS;
}

VkResult vkGetPhysicalDeviceSurfaceFormatsKHR(VkPhysicalDevice physicalDevice, VkSurfaceKHR surface, uint32_t *pSurfaceFormatCount, VkSurfaceFormatKHR *pSurfaceFormats) {
    (void)physicalDevice;
    (void)surface;
    static const VkSurfaceFormatKHR formats[] = {
        {.format = VK_FORMAT_B8G8R8A8_SRGB, .colorSpace = VK_COLOR_SPACE_SRGB_NONLINEAR_KHR},
        {.format = VK_FORMAT_B8G8R8A8_UNORM, .colorSpace = VK_COLOR_SPACE_SRGB_NONLINEAR_KHR},
    };
    if (!pSurfaceFormats) {
        *pSurfaceFormatCount = 2;
        return VK_SUCCESS;
    }
    uint32_t n = *pSurfaceFormatCount < 2 ? *pSurfaceFormatCount : 2;
    memcpy(pSurfaceFormats, formats, n * sizeof(formats[0]));
    *pSurfaceFormatCount = n;
    return VK_SUCCESS;
}

VkResult vkGetPhysicalDeviceSurfacePresentModesKHR(VkPhysicalDevice physicalDevice, VkSurfaceKHR surface, uint32_t *pPresentModeCount, VkPresentModeKHR *pPresentModes) {
    (void)physicalDevice;
    (void)surface;
    static const VkPresentModeKHR modes[] = {VK_PRESENT_MODE_FIFO_KHR, VK_PRESENT_MODE_MAILBOX_KHR};
    if (!pPresentModes) {
        *pPresentModeCount = 2;
        return VK_SUCCESS;
    }
    uint32_t n = *pPresentModeCount < 2 ? *pPresentModeCount : 2;
    memcpy(pPresentModes, modes, n * sizeof(modes[0]));
    *pPresentModeCount = n;
    return VK_SUCCESS;
}

VkResult vkGetPhysicalDeviceSurfaceCapabilitiesKHR(VkPhysicalDevice physicalDevice, VkSurfaceKHR surface, VkSurfaceCapabilitiesKHR *pSurfaceCapabilities) {
    (void)physicalDevice;
    struct VkSurfaceKHR_T *surf = ND_PTR(struct VkSurfaceKHR_T, surface);
    memset(pSurfaceCapabilities, 0, sizeof(*pSurfaceCapabilities));
    pSurfaceCapabilities->minImageCount = 2;
    pSurfaceCapabilities->maxImageCount = 3;
    pSurfaceCapabilities->currentExtent.width = surf ? surf->width : 800;
    pSurfaceCapabilities->currentExtent.height = surf ? surf->height : 600;
    pSurfaceCapabilities->minImageExtent.width = 1;
    pSurfaceCapabilities->minImageExtent.height = 1;
    pSurfaceCapabilities->maxImageExtent.width = 8192;
    pSurfaceCapabilities->maxImageExtent.height = 8192;
    pSurfaceCapabilities->maxImageArrayLayers = 1;
    pSurfaceCapabilities->supportedTransforms = VK_SURFACE_TRANSFORM_IDENTITY_BIT_KHR;
    pSurfaceCapabilities->currentTransform = VK_SURFACE_TRANSFORM_IDENTITY_BIT_KHR;
    pSurfaceCapabilities->supportedCompositeAlpha = VK_COMPOSITE_ALPHA_OPAQUE_BIT_KHR;
    pSurfaceCapabilities->supportedUsageFlags = VK_IMAGE_USAGE_COLOR_ATTACHMENT_BIT;
    return VK_SUCCESS;
}

VkResult glfwCreateWindowSurface(VkInstance instance, GLFWwindow *window, const VkAllocationCallbacks *allocator, VkSurfaceKHR *surface) {
    (void)allocator;
    if (!instance || !instance->wgpu || !surface) {
        return VK_ERROR_INITIALIZATION_FAILED;
    }
    struct VkSurfaceKHR_T *surf = calloc(1, sizeof(*surf));
    if (!surf) {
        return VK_ERROR_OUT_OF_HOST_MEMORY;
    }
    int w = 800, h = 600;
    glfwGetFramebufferSize(window, &w, &h);
    surf->width = (uint32_t)w;
    surf->height = (uint32_t)h;

    WGPUSurfaceDescriptorFromCanvasHTMLSelector canvas = {
        .chain = {.sType = WGPUSType_SurfaceDescriptorFromCanvasHTMLSelector},
        .selector = "#canvas",
    };
    WGPUSurfaceDescriptor desc = {
        .nextInChain = &canvas.chain,
        .label = "canvas",
    };
    surf->wgpu = wgpuInstanceCreateSurface(instance->wgpu, &desc);
    if (!surf->wgpu) {
        free(surf);
        return VK_ERROR_INITIALIZATION_FAILED;
    }
    *surface = (VkSurfaceKHR)ND_HANDLE(surf);
    return VK_SUCCESS;
}

VkResult vkCreateDevice(VkPhysicalDevice physicalDevice, const VkDeviceCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkDevice *pDevice) {
    (void)pCreateInfo;
    (void)pAllocator;
    struct VkDevice_T *dev = calloc(1, sizeof(*dev));
    struct VkQueue_T *queue = calloc(1, sizeof(*queue));
    if (!dev || !queue) {
        free(dev);
        free(queue);
        return VK_ERROR_OUT_OF_HOST_MEMORY;
    }
    dev->wgpu = emscripten_webgpu_get_device();
    if (!dev->wgpu) {
        free(dev);
        free(queue);
        return VK_ERROR_INITIALIZATION_FAILED;
    }
    wgpuDeviceSetUncapturedErrorCallback(dev->wgpu, on_wgpu_error, NULL);
    dev->queue = wgpuDeviceGetQueue(dev->wgpu);
    queue->wgpu = dev->queue;
    dev->vk_queue = queue;
    dev->phys = physicalDevice;
    *pDevice = dev;
    return VK_SUCCESS;
}

void vkDestroyDevice(VkDevice device, const VkAllocationCallbacks *pAllocator) {
    (void)pAllocator;
    if (!device) {
        return;
    }
    free(device->vk_queue);
    if (device->wgpu) {
        wgpuDeviceRelease(device->wgpu);
    }
    free(device);
}

void vkGetDeviceQueue(VkDevice device, uint32_t queueFamilyIndex, uint32_t queueIndex, VkQueue *pQueue) {
    (void)queueFamilyIndex;
    (void)queueIndex;
    *pQueue = device->vk_queue;
}

VkResult vkCreateSwapchainKHR(VkDevice device, const VkSwapchainCreateInfoKHR *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkSwapchainKHR *pSwapchain) {
    (void)pAllocator;
    struct VkSurfaceKHR_T *surf = ND_PTR(struct VkSurfaceKHR_T, pCreateInfo->surface);
    struct VkSwapchainKHR_T *sc = calloc(1, sizeof(*sc));
    if (!sc) {
        return VK_ERROR_OUT_OF_HOST_MEMORY;
    }
    sc->device = device;
    sc->surface = surf;
    sc->vk_format = pCreateInfo->imageFormat;
    sc->wgpu_format = canvas_format();
    sc->width = pCreateInfo->imageExtent.width;
    sc->height = pCreateInfo->imageExtent.height;
    sc->image_count = pCreateInfo->minImageCount;
    if (sc->image_count == 0) {
        sc->image_count = 1;
    }
    if (sc->image_count > MAX_SWAP_IMAGES) {
        sc->image_count = MAX_SWAP_IMAGES;
    }
    for (uint32_t i = 0; i < sc->image_count; i++) {
        sc->images[i] = calloc(1, sizeof(*sc->images[i]));
        if (!sc->images[i]) {
            return VK_ERROR_OUT_OF_HOST_MEMORY;
        }
    }

    WGPUSurfaceConfiguration config = {
        .device = device->wgpu,
        .format = sc->wgpu_format,
        .usage = WGPUTextureUsage_RenderAttachment,
        .alphaMode = WGPUCompositeAlphaMode_Auto,
        .width = sc->width,
        .height = sc->height,
        .presentMode = WGPUPresentMode_Fifo,
    };
    wgpuSurfaceConfigure(surf->wgpu, &config);
    g_active_swapchain = sc;
    *pSwapchain = (VkSwapchainKHR)ND_HANDLE(sc);
    return VK_SUCCESS;
}

void vkDestroySwapchainKHR(VkDevice device, VkSwapchainKHR swapchain, const VkAllocationCallbacks *pAllocator) {
    (void)device;
    (void)pAllocator;
    struct VkSwapchainKHR_T *sc = ND_PTR(struct VkSwapchainKHR_T, swapchain);
    if (!sc) {
        return;
    }
    for (uint32_t i = 0; i < sc->image_count; i++) {
        free(sc->images[i]);
    }
    if (g_active_swapchain == sc) {
        g_active_swapchain = NULL;
    }
    free(sc);
}

VkResult vkGetSwapchainImagesKHR(VkDevice device, VkSwapchainKHR swapchain, uint32_t *pSwapchainImageCount, VkImage *pSwapchainImages) {
    (void)device;
    struct VkSwapchainKHR_T *sc = ND_PTR(struct VkSwapchainKHR_T, swapchain);
    if (!pSwapchainImages) {
        *pSwapchainImageCount = sc->image_count;
        return VK_SUCCESS;
    }
    uint32_t n = *pSwapchainImageCount < sc->image_count ? *pSwapchainImageCount : sc->image_count;
    for (uint32_t i = 0; i < n; i++) {
        pSwapchainImages[i] = (VkImage)ND_HANDLE(sc->images[i]);
    }
    *pSwapchainImageCount = n;
    return VK_SUCCESS;
}

VkResult vkCreateImageView(VkDevice device, const VkImageViewCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkImageView *pView) {
    (void)device;
    (void)pAllocator;
    struct VkImageView_T *view = calloc(1, sizeof(*view));
    if (!view) {
        return VK_ERROR_OUT_OF_HOST_MEMORY;
    }
    view->image = ND_PTR(struct VkImage_T, pCreateInfo->image);
    if (view->image) {
        view->image->view = view;
    }
    *pView = (VkImageView)ND_HANDLE(view);
    return VK_SUCCESS;
}

void vkDestroyImageView(VkDevice device, VkImageView imageView, const VkAllocationCallbacks *pAllocator) {
    (void)device;
    (void)pAllocator;
    struct VkImageView_T *view = ND_PTR(struct VkImageView_T, imageView);
    if (!view) {
        return;
    }
    if (view->wgpu) {
        wgpuTextureViewRelease(view->wgpu);
    }
    if (view->image && view->image->view == view) {
        view->image->view = NULL;
    }
    free(view);
}

VkResult vkCreateShaderModule(VkDevice device, const VkShaderModuleCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkShaderModule *pShaderModule) {
    (void)pAllocator;
    struct VkShaderModule_T *mod = calloc(1, sizeof(*mod));
    if (!mod) {
        return VK_ERROR_OUT_OF_HOST_MEMORY;
    }
    int model = spirv_execution_model(pCreateInfo->pCode, pCreateInfo->codeSize);
    if (model < 0) {
        model = (g_shader_seq++ == 0) ? 0 : 4;
    }
    mod->is_vertex = (model == 0);
    const char *wgsl = mod->is_vertex ? kVertexWGSL : kFragmentWGSL;
    mod->wgpu = create_wgsl_module(device->wgpu, wgsl);
    if (!mod->wgpu) {
        free(mod);
        return VK_ERROR_INITIALIZATION_FAILED;
    }
    *pShaderModule = (VkShaderModule)ND_HANDLE(mod);
    return VK_SUCCESS;
}

void vkDestroyShaderModule(VkDevice device, VkShaderModule shaderModule, const VkAllocationCallbacks *pAllocator) {
    (void)device;
    (void)pAllocator;
    struct VkShaderModule_T *mod = ND_PTR(struct VkShaderModule_T, shaderModule);
    if (!mod) {
        return;
    }
    if (mod->wgpu) {
        wgpuShaderModuleRelease(mod->wgpu);
    }
    free(mod);
}

VkResult vkCreatePipelineLayout(VkDevice device, const VkPipelineLayoutCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkPipelineLayout *pPipelineLayout) {
    (void)device;
    (void)pCreateInfo;
    (void)pAllocator;
    struct VkPipelineLayout_T *layout = calloc(1, sizeof(*layout));
    if (!layout) {
        return VK_ERROR_OUT_OF_HOST_MEMORY;
    }
    *pPipelineLayout = (VkPipelineLayout)ND_HANDLE(layout);
    return VK_SUCCESS;
}

void vkDestroyPipelineLayout(VkDevice device, VkPipelineLayout pipelineLayout, const VkAllocationCallbacks *pAllocator) {
    (void)device;
    (void)pAllocator;
    free(ND_PTR(struct VkPipelineLayout_T, pipelineLayout));
}

VkResult vkCreateRenderPass(VkDevice device, const VkRenderPassCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkRenderPass *pRenderPass) {
    (void)device;
    (void)pAllocator;
    struct VkRenderPass_T *rp = calloc(1, sizeof(*rp));
    if (!rp) {
        return VK_ERROR_OUT_OF_HOST_MEMORY;
    }
    rp->format = VK_FORMAT_B8G8R8A8_SRGB;
    if (pCreateInfo && pCreateInfo->attachmentCount > 0 && pCreateInfo->pAttachments) {
        rp->format = pCreateInfo->pAttachments[0].format;
    }
    *pRenderPass = (VkRenderPass)ND_HANDLE(rp);
    return VK_SUCCESS;
}

void vkDestroyRenderPass(VkDevice device, VkRenderPass renderPass, const VkAllocationCallbacks *pAllocator) {
    (void)device;
    (void)pAllocator;
    free(ND_PTR(struct VkRenderPass_T, renderPass));
}

VkResult vkCreateGraphicsPipelines(VkDevice device, VkPipelineCache pipelineCache, uint32_t createInfoCount, const VkGraphicsPipelineCreateInfo *pCreateInfos, const VkAllocationCallbacks *pAllocator, VkPipeline *pPipelines) {
    (void)pipelineCache;
    (void)pAllocator;
    for (uint32_t i = 0; i < createInfoCount; i++) {
        const VkGraphicsPipelineCreateInfo *ci = &pCreateInfos[i];
        struct VkShaderModule_T *vs = NULL;
        struct VkShaderModule_T *fs = NULL;
        for (uint32_t s = 0; s < ci->stageCount; s++) {
            struct VkShaderModule_T *mod = ND_PTR(struct VkShaderModule_T, ci->pStages[s].module);
            if (ci->pStages[s].stage & VK_SHADER_STAGE_VERTEX_BIT) {
                vs = mod;
            }
            if (ci->pStages[s].stage & VK_SHADER_STAGE_FRAGMENT_BIT) {
                fs = mod;
            }
        }
        if (!vs || !fs || !vs->wgpu || !fs->wgpu) {
            return VK_ERROR_INITIALIZATION_FAILED;
        }

        WGPUTextureFormat format = canvas_format();

        WGPUColorTargetState colorTarget = {
            .format = format,
            .writeMask = WGPUColorWriteMask_All,
        };
        WGPUFragmentState fragment = {
            .module = fs->wgpu,
            .entryPoint = "main",
            .targetCount = 1,
            .targets = &colorTarget,
        };
        WGPURenderPipelineDescriptor desc = {
            .vertex = {
                .module = vs->wgpu,
                .entryPoint = "main",
            },
            .primitive = {
                .topology = WGPUPrimitiveTopology_TriangleList,
                .frontFace = WGPUFrontFace_CW,
                .cullMode = WGPUCullMode_Back,
            },
            .multisample = {
                .count = 1,
                .mask = 0xFFFFFFFFu,
            },
            .fragment = &fragment,
        };
        struct VkPipeline_T *pipe = calloc(1, sizeof(*pipe));
        if (!pipe) {
            return VK_ERROR_OUT_OF_HOST_MEMORY;
        }
        pipe->wgpu = wgpuDeviceCreateRenderPipeline(device->wgpu, &desc);
        if (!pipe->wgpu) {
            free(pipe);
            return VK_ERROR_INITIALIZATION_FAILED;
        }
        pPipelines[i] = (VkPipeline)ND_HANDLE(pipe);
    }
    return VK_SUCCESS;
}

void vkDestroyPipeline(VkDevice device, VkPipeline pipeline, const VkAllocationCallbacks *pAllocator) {
    (void)device;
    (void)pAllocator;
    struct VkPipeline_T *pipe = ND_PTR(struct VkPipeline_T, pipeline);
    if (!pipe) {
        return;
    }
    if (pipe->wgpu) {
        wgpuRenderPipelineRelease(pipe->wgpu);
    }
    free(pipe);
}

VkResult vkCreateFramebuffer(VkDevice device, const VkFramebufferCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkFramebuffer *pFramebuffer) {
    (void)device;
    (void)pAllocator;
    struct VkFramebuffer_T *fb = calloc(1, sizeof(*fb));
    if (!fb) {
        return VK_ERROR_OUT_OF_HOST_MEMORY;
    }
    if (pCreateInfo->attachmentCount > 0 && pCreateInfo->pAttachments) {
        fb->color = ND_PTR(struct VkImageView_T, pCreateInfo->pAttachments[0]);
    }
    fb->width = pCreateInfo->width;
    fb->height = pCreateInfo->height;
    *pFramebuffer = (VkFramebuffer)ND_HANDLE(fb);
    return VK_SUCCESS;
}

void vkDestroyFramebuffer(VkDevice device, VkFramebuffer framebuffer, const VkAllocationCallbacks *pAllocator) {
    (void)device;
    (void)pAllocator;
    free(ND_PTR(struct VkFramebuffer_T, framebuffer));
}

VkResult vkCreateCommandPool(VkDevice device, const VkCommandPoolCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkCommandPool *pCommandPool) {
    (void)pCreateInfo;
    (void)pAllocator;
    struct VkCommandPool_T *pool = calloc(1, sizeof(*pool));
    if (!pool) {
        return VK_ERROR_OUT_OF_HOST_MEMORY;
    }
    pool->device = device;
    *pCommandPool = (VkCommandPool)ND_HANDLE(pool);
    return VK_SUCCESS;
}

void vkDestroyCommandPool(VkDevice device, VkCommandPool commandPool, const VkAllocationCallbacks *pAllocator) {
    (void)device;
    (void)pAllocator;
    free(ND_PTR(struct VkCommandPool_T, commandPool));
}

VkResult vkAllocateCommandBuffers(VkDevice device, const VkCommandBufferAllocateInfo *pAllocateInfo, VkCommandBuffer *pCommandBuffers) {
    for (uint32_t i = 0; i < pAllocateInfo->commandBufferCount; i++) {
        struct VkCommandBuffer_T *cb = calloc(1, sizeof(*cb));
        if (!cb) {
            return VK_ERROR_OUT_OF_HOST_MEMORY;
        }
        cb->device = device;
        pCommandBuffers[i] = cb;
    }
    return VK_SUCCESS;
}

VkResult vkBeginCommandBuffer(VkCommandBuffer commandBuffer, const VkCommandBufferBeginInfo *pBeginInfo) {
    (void)pBeginInfo;
    commandBuffer->framebuffer = NULL;
    commandBuffer->pipeline = NULL;
    commandBuffer->has_viewport = 0;
    commandBuffer->has_scissor = 0;
    commandBuffer->draw_vertices = 0;
    return VK_SUCCESS;
}

VkResult vkEndCommandBuffer(VkCommandBuffer commandBuffer) {
    (void)commandBuffer;
    return VK_SUCCESS;
}

VkResult vkResetCommandBuffer(VkCommandBuffer commandBuffer, VkCommandBufferResetFlags flags) {
    (void)flags;
    return vkBeginCommandBuffer(commandBuffer, NULL);
}

void vkCmdBeginRenderPass(VkCommandBuffer commandBuffer, const VkRenderPassBeginInfo *pRenderPassBegin, VkSubpassContents contents) {
    (void)contents;
    commandBuffer->framebuffer = ND_PTR(struct VkFramebuffer_T, pRenderPassBegin->framebuffer);
}

void vkCmdEndRenderPass(VkCommandBuffer commandBuffer) {
    (void)commandBuffer;
}

void vkCmdBindPipeline(VkCommandBuffer commandBuffer, VkPipelineBindPoint pipelineBindPoint, VkPipeline pipeline) {
    (void)pipelineBindPoint;
    commandBuffer->pipeline = ND_PTR(struct VkPipeline_T, pipeline);
}

void vkCmdSetViewport(VkCommandBuffer commandBuffer, uint32_t firstViewport, uint32_t viewportCount, const VkViewport *pViewports) {
    (void)firstViewport;
    if (viewportCount > 0 && pViewports) {
        commandBuffer->viewport = pViewports[0];
        commandBuffer->has_viewport = 1;
    }
}

void vkCmdSetScissor(VkCommandBuffer commandBuffer, uint32_t firstScissor, uint32_t scissorCount, const VkRect2D *pScissors) {
    (void)firstScissor;
    if (scissorCount > 0 && pScissors) {
        commandBuffer->scissor = pScissors[0];
        commandBuffer->has_scissor = 1;
    }
}

void vkCmdDraw(VkCommandBuffer commandBuffer, uint32_t vertexCount, uint32_t instanceCount, uint32_t firstVertex, uint32_t firstInstance) {
    (void)instanceCount;
    (void)firstVertex;
    (void)firstInstance;
    commandBuffer->draw_vertices = vertexCount;
}

VkResult vkCreateSemaphore(VkDevice device, const VkSemaphoreCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkSemaphore *pSemaphore) {
    (void)device;
    (void)pCreateInfo;
    (void)pAllocator;
    struct VkSemaphore_T *sem = calloc(1, sizeof(*sem));
    if (!sem) {
        return VK_ERROR_OUT_OF_HOST_MEMORY;
    }
    *pSemaphore = (VkSemaphore)ND_HANDLE(sem);
    return VK_SUCCESS;
}

void vkDestroySemaphore(VkDevice device, VkSemaphore semaphore, const VkAllocationCallbacks *pAllocator) {
    (void)device;
    (void)pAllocator;
    free(ND_PTR(struct VkSemaphore_T, semaphore));
}

VkResult vkCreateFence(VkDevice device, const VkFenceCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkFence *pFence) {
    (void)device;
    (void)pAllocator;
    struct VkFence_T *fence = calloc(1, sizeof(*fence));
    if (!fence) {
        return VK_ERROR_OUT_OF_HOST_MEMORY;
    }
    fence->signaled = (pCreateInfo && (pCreateInfo->flags & VK_FENCE_CREATE_SIGNALED_BIT)) ? 1 : 0;
    *pFence = (VkFence)ND_HANDLE(fence);
    return VK_SUCCESS;
}

void vkDestroyFence(VkDevice device, VkFence fence, const VkAllocationCallbacks *pAllocator) {
    (void)device;
    (void)pAllocator;
    free(ND_PTR(struct VkFence_T, fence));
}

VkResult vkWaitForFences(VkDevice device, uint32_t fenceCount, const VkFence *pFences, VkBool32 waitAll, uint64_t timeout) {
    (void)device;
    (void)waitAll;
    (void)timeout;
    for (uint32_t i = 0; i < fenceCount; i++) {
        struct VkFence_T *f = ND_PTR(struct VkFence_T, pFences[i]);
        if (f) {
            f->signaled = 1;
        }
    }
    return VK_SUCCESS;
}

VkResult vkResetFences(VkDevice device, uint32_t fenceCount, const VkFence *pFences) {
    (void)device;
    for (uint32_t i = 0; i < fenceCount; i++) {
        struct VkFence_T *f = ND_PTR(struct VkFence_T, pFences[i]);
        if (f) {
            f->signaled = 0;
        }
    }
    return VK_SUCCESS;
}

VkResult vkAcquireNextImageKHR(VkDevice device, VkSwapchainKHR swapchain, uint64_t timeout, VkSemaphore semaphore, VkFence fence, uint32_t *pImageIndex) {
    (void)device;
    (void)timeout;
    (void)semaphore;
    (void)fence;
    struct VkSwapchainKHR_T *sc = ND_PTR(struct VkSwapchainKHR_T, swapchain);
    WGPUSurfaceTexture st = {0};
    wgpuSurfaceGetCurrentTexture(sc->surface->wgpu, &st);
    if (st.status != WGPUSurfaceGetCurrentTextureStatus_Success || !st.texture) {
        fprintf(stderr, "webgpu: getCurrentTexture status=%u\n", (unsigned)st.status);
        return VK_ERROR_OUT_OF_DATE_KHR;
    }
    if (sc->current_texture && sc->current_texture != st.texture) {
        wgpuTextureRelease(sc->current_texture);
    }
    sc->current_texture = st.texture;
    sc->images[0]->texture = st.texture;
    refresh_image_view(sc->images[0]);
    *pImageIndex = 0;
    return VK_SUCCESS;
}

static void execute_command_buffer(struct VkCommandBuffer_T *cb) {
    if (!cb || !cb->device || !cb->pipeline || !cb->pipeline->wgpu || !cb->framebuffer || !cb->framebuffer->color || !cb->framebuffer->color->wgpu) {
        fprintf(stderr, "webgpu: command buffer missing render state\n");
        return;
    }
    WGPUCommandEncoder encoder = wgpuDeviceCreateCommandEncoder(cb->device->wgpu, NULL);
    WGPURenderPassColorAttachment color = {
        .view = cb->framebuffer->color->wgpu,
        .depthSlice = WGPU_DEPTH_SLICE_UNDEFINED,
        .loadOp = WGPULoadOp_Clear,
        .storeOp = WGPUStoreOp_Store,
        .clearValue = {0.0, 0.0, 0.0, 1.0},
    };
    WGPURenderPassDescriptor pass_desc = {
        .colorAttachmentCount = 1,
        .colorAttachments = &color,
    };
    WGPURenderPassEncoder pass = wgpuCommandEncoderBeginRenderPass(encoder, &pass_desc);
    wgpuRenderPassEncoderSetPipeline(pass, cb->pipeline->wgpu);
    if (cb->has_viewport) {
        wgpuRenderPassEncoderSetViewport(
            pass,
            cb->viewport.x,
            cb->viewport.y,
            cb->viewport.width,
            cb->viewport.height,
            cb->viewport.minDepth,
            cb->viewport.maxDepth);
    }
    if (cb->has_scissor) {
        wgpuRenderPassEncoderSetScissorRect(
            pass,
            (uint32_t)cb->scissor.offset.x,
            (uint32_t)cb->scissor.offset.y,
            cb->scissor.extent.width,
            cb->scissor.extent.height);
    }
    wgpuRenderPassEncoderDraw(pass, cb->draw_vertices ? cb->draw_vertices : 3, 1, 0, 0);
    wgpuRenderPassEncoderEnd(pass);
    WGPUCommandBuffer cmd = wgpuCommandEncoderFinish(encoder, NULL);
    wgpuQueueSubmit(cb->device->queue, 1, &cmd);
    wgpuCommandBufferRelease(cmd);
    wgpuRenderPassEncoderRelease(pass);
    wgpuCommandEncoderRelease(encoder);
}

VkResult vkQueueSubmit(VkQueue queue, uint32_t submitCount, const VkSubmitInfo *pSubmits, VkFence fence) {
    (void)queue;
    for (uint32_t s = 0; s < submitCount; s++) {
        for (uint32_t c = 0; c < pSubmits[s].commandBufferCount; c++) {
            execute_command_buffer(pSubmits[s].pCommandBuffers[c]);
        }
    }
    if (fence) {
        struct VkFence_T *f = ND_PTR(struct VkFence_T, fence);
        if (f) {
            f->signaled = 1;
        }
    }
    return VK_SUCCESS;
}

VkResult vkQueuePresentKHR(VkQueue queue, const VkPresentInfoKHR *pPresentInfo) {
    (void)queue;
    (void)pPresentInfo;
    /* Emscripten presents the canvas on rAF; wgpuSurfacePresent aborts. */
    return VK_SUCCESS;
}

VkResult vkDeviceWaitIdle(VkDevice device) {
    (void)device;
    return VK_SUCCESS;
}

void vkDestroySurfaceKHR(VkInstance instance, VkSurfaceKHR surface, const VkAllocationCallbacks *pAllocator) {
    (void)instance;
    (void)pAllocator;
    struct VkSurfaceKHR_T *surf = ND_PTR(struct VkSurfaceKHR_T, surface);
    if (!surf) {
        return;
    }
    if (surf->wgpu) {
        wgpuSurfaceRelease(surf->wgpu);
    }
    free(surf);
}
