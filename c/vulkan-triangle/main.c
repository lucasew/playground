#include<stdio.h>
#include<stdlib.h>
#include<string.h>
#include<limits.h>
#include<stdint.h>
#include<errno.h>
#include<fcntl.h>
#include<sys/stat.h>
#include<sys/mman.h>
#include<unistd.h>

#define GLFW_INCLUDE_VULKAN

#include<GLFW/glfw3.h>

const char* APP_NAME = "V U L K A N";
const char* const VULKAN_VALIDATION_LAYERS[] = {
    "VK_LAYER_KHRONOS_validation"
};
const char * const DEVICE_EXTENSIONS[] = {
    "VK_KHR_swapchain"
};


VkInstance instance;
VkDebugUtilsMessengerCreateInfoEXT debugCreateInfo;
VkDebugUtilsMessengerEXT debugMessenger;
VkSurfaceKHR surface;
VkPhysicalDevice physicalDevice;
VkDevice device;
VkQueue graphicsQueue;
VkQueue presentQueue;
VkSurfaceCapabilitiesKHR surfaceCapatibilitesDetails;
VkExtent2D swapChainExtent;
VkSurfaceFormatKHR surfaceFormat;
VkPresentModeKHR presentMode;
VkSwapchainKHR swapChain;
VkImage* swapchainImages = NULL;
VkImageView* swapchainImageViews = NULL;
VkFormat swapChainImageFormat;
VkShaderModule vertexShaderModule;
VkShaderModule fragmentShaderModule;
VkPipelineLayout pipelineLayout;
VkRenderPass renderPass;
VkPipeline graphicsPipeline;
VkFramebuffer *swapChainFramebuffers = NULL;
VkCommandPool commandPool;
VkCommandBuffer commandBuffer;

VkSemaphore imageAvailableSemaphore;
VkSemaphore renderFinishedSemaphore;
VkFence inFlightFence;

#define clamp(x, lo, hi) (x < lo ? lo : x > hi ? hi : x)

static void glfwErrorCallback(int error, const char* description) {
    fprintf(stderr, "glfwdebug %i: %s\n", error, description);
}

static VKAPI_ATTR VkBool32 VKAPI_CALL vulkanDebugCallback(
        VkDebugUtilsMessageSeverityFlagBitsEXT severity,
        VkDebugUtilsMessageTypeFlagsEXT type,
        const VkDebugUtilsMessengerCallbackDataEXT* pCallbackData,
        void* pUserData) {
    fprintf(stderr, "vkdebug %i %i: %s\n", severity, type, pCallbackData->pMessage);
    return VK_FALSE;
}

VkResult readShader(VkDevice device, VkShaderModule *shaderModule, const char* filename) {
    struct stat sb;
    int fd = open(filename, O_RDONLY);
    if (fd == -1) {
        fprintf(stderr, "can't open file '%s': %s", filename, strerror(errno));
        return VK_ERROR_MEMORY_MAP_FAILED;
    }
    if (fstat(fd, &sb) == -1) {
        fprintf(stderr, "can't fstat file '%s': %s", filename, strerror(errno));
        return VK_ERROR_MEMORY_MAP_FAILED;
    }
    const uint32_t* addr = mmap(NULL, sb.st_size, PROT_READ, MAP_PRIVATE, fd, 0);
    if (addr == MAP_FAILED) {
        fprintf(stderr, "can't mmap file '%s': %s", filename, strerror(errno));
        return VK_ERROR_MEMORY_MAP_FAILED;
    }
    VkShaderModuleCreateInfo createInfo;
    createInfo.sType = VK_STRUCTURE_TYPE_SHADER_MODULE_CREATE_INFO;
    createInfo.codeSize = sb.st_size;
    createInfo.pCode = addr;
    close(fd);
    return vkCreateShaderModule(device, &createInfo, NULL, shaderModule);
}


void showExtensions() {
    // Demonstração das extensões
    uint32_t extensionCount = 0;
    vkEnumerateInstanceExtensionProperties(NULL, &extensionCount, NULL);
    fprintf(stderr, "Number of vulkan extensions supported: %i\n", extensionCount);
    VkExtensionProperties* extensionProperties = malloc(sizeof(VkExtensionProperties)*extensionCount);
    vkEnumerateInstanceExtensionProperties(NULL, &extensionCount, extensionProperties);
    for (int i = 0; i < extensionCount; i++) {
        VkExtensionProperties property = extensionProperties[i];
        fprintf(stderr, "\tExtension: %s (%i)\n", property.extensionName, property.specVersion);
    }
    free(extensionProperties);
}

void showValidationLayers() {
    uint32_t validationLayersCount;
    vkEnumerateInstanceLayerProperties(&validationLayersCount, NULL);
    VkLayerProperties* validationLayers = malloc(sizeof(VkLayerProperties)*validationLayersCount);
    vkEnumerateInstanceLayerProperties(&validationLayersCount, validationLayers);
    fprintf(stderr, "Number of validation layers supported: %i\n", validationLayersCount);
    for (int i = 0; i < validationLayersCount; i++) {
        VkLayerProperties layer = validationLayers[i];
        fprintf(stderr, "\t Layer: %s (spec:%i, impl=%i): %s\n", layer.layerName, layer.specVersion, layer.implementationVersion, layer.description);
    }
}

void getUsedValidationLayers(VkInstanceCreateInfo *createInfo) {
    uint32_t validationLayersCount;
    vkEnumerateInstanceLayerProperties(&validationLayersCount, NULL);
    VkLayerProperties* validationLayers = malloc(sizeof(VkLayerProperties)*validationLayersCount);
    vkEnumerateInstanceLayerProperties(&validationLayersCount, validationLayers);
    const char **usedValidationLayers = malloc(sizeof(char*)*validationLayersCount);
    for (int i = 0; i < validationLayersCount; i++) {
        usedValidationLayers[i] = validationLayers[i].layerName;
    }
    createInfo->enabledLayerCount = validationLayersCount;
    createInfo->ppEnabledLayerNames = usedValidationLayers;
}

void getUsedExtensions(VkInstanceCreateInfo *createInfo) {
    uint32_t glfwExtensionCount;
    const char **glfwExtensions = glfwGetRequiredInstanceExtensions(&glfwExtensionCount);

    const char **vulkanExtensions = malloc(sizeof(char*)*(glfwExtensionCount+1));
    memcpy(vulkanExtensions, glfwExtensions, sizeof(char*)*glfwExtensionCount);
    vulkanExtensions[glfwExtensionCount] = VK_EXT_DEBUG_UTILS_EXTENSION_NAME;

    createInfo->enabledExtensionCount = glfwExtensionCount + 1;
    createInfo->ppEnabledExtensionNames = vulkanExtensions;
}
VkResult setupDebug(VkInstance instance, VkDebugUtilsMessengerCreateInfoEXT *debugCreateInfo, VkDebugUtilsMessengerEXT *debugMessenger) {
    debugCreateInfo->sType = VK_STRUCTURE_TYPE_DEBUG_UTILS_MESSENGER_CREATE_INFO_EXT;
    debugCreateInfo->messageSeverity =
          VK_DEBUG_UTILS_MESSAGE_SEVERITY_VERBOSE_BIT_EXT
        | VK_DEBUG_UTILS_MESSAGE_SEVERITY_WARNING_BIT_EXT
        | VK_DEBUG_UTILS_MESSAGE_TYPE_PERFORMANCE_BIT_EXT;

    debugCreateInfo->pfnUserCallback = vulkanDebugCallback;
    debugCreateInfo->pUserData = NULL;

    PFN_vkCreateDebugUtilsMessengerEXT handler = (PFN_vkCreateDebugUtilsMessengerEXT) vkGetInstanceProcAddr(instance, "vkCreateDebugUtilsMessengerEXT");
    if (handler) {
        return handler(instance, debugCreateInfo, NULL, debugMessenger);
    }
    fprintf(stderr, "setupDebug: vkCreateDebugUtilsMessengerEXT not found\n");
    return VK_ERROR_EXTENSION_NOT_PRESENT;
}

int getFirstQueueFamilyOfType(VkPhysicalDevice device, VkQueueFlags flag) {
    uint32_t queueFamilyCount = 0;
    vkGetPhysicalDeviceQueueFamilyProperties(device, &queueFamilyCount, NULL);
    VkQueueFamilyProperties* queueFamilies = malloc(sizeof(VkQueueFamilyProperties)*queueFamilyCount);
    vkGetPhysicalDeviceQueueFamilyProperties(device, &queueFamilyCount, queueFamilies);

    int ret = -1; // nothing found
    for (int i = 0; i < queueFamilyCount; i++) {
        VkQueueFamilyProperties queueFamily = queueFamilies[i];
        if (queueFamily.queueFlags & flag) {
            ret = i;
        }
    }
    free(queueFamilies);
    return ret;
}

VkSurfaceFormatKHR getSwapSurfaceFormat(VkPhysicalDevice device, VkSurfaceKHR surface) {
    uint32_t deviceSurfaceFormatsCount;
    vkGetPhysicalDeviceSurfaceFormatsKHR(device, surface, &deviceSurfaceFormatsCount, NULL);
    VkSurfaceFormatKHR* formats = malloc(sizeof(VkSurfaceFormatKHR)*deviceSurfaceFormatsCount);
    vkGetPhysicalDeviceSurfaceFormatsKHR(device, surface, &deviceSurfaceFormatsCount, formats);
    VkSurfaceFormatKHR ret;
    for (int i = 0; i < deviceSurfaceFormatsCount; i++) {
        VkSurfaceFormatKHR format = formats[i];
        if (format.format == VK_FORMAT_B8G8R8A8_SRGB && format.colorSpace == VK_COLOR_SPACE_SRGB_NONLINEAR_KHR) {
            ret = format;
            break;
        }
    }
    free(formats);
    return ret;
}

VkPresentModeKHR getSwapPresentMode(VkPhysicalDevice device, VkSurfaceKHR surface) {
    uint32_t devicePresentModeCount;
    vkGetPhysicalDeviceSurfacePresentModesKHR(device, surface, &devicePresentModeCount, NULL);
    VkPresentModeKHR ret = VK_PRESENT_MODE_FIFO_KHR;
    VkPresentModeKHR* modes = malloc(sizeof(VkPresentModeKHR)*devicePresentModeCount);
    vkGetPhysicalDeviceSurfacePresentModesKHR(device, surface, &devicePresentModeCount, modes);
    for (int i = 0; i < devicePresentModeCount; i++) {
        VkPresentModeKHR presentMode = modes[i];
        if (presentMode == VK_PRESENT_MODE_MAILBOX_KHR) {
            ret = presentMode;
        }
    }
    free(modes);
    return ret;

}


VkPhysicalDevice getDevice(VkInstance instance, VkSurfaceKHR surface) {
    uint32_t deviceCount = 0;
    vkEnumeratePhysicalDevices(instance, &deviceCount, NULL);
    if (deviceCount == 0) {
        return VK_NULL_HANDLE;
    }
    VkPhysicalDevice* devices = malloc(sizeof(VkPhysicalDevice)*deviceCount);
    vkEnumeratePhysicalDevices(instance, &deviceCount, devices);
    fprintf(stderr, "Number of devices supported: %i\n", deviceCount);
    VkPhysicalDevice chosenDevice = VK_NULL_HANDLE;
    int best_score = 0;

    for (int i = 0; i < deviceCount; i++) {
        VkPhysicalDevice device = devices[i];
        VkPhysicalDeviceProperties deviceProperties;
        vkGetPhysicalDeviceProperties(device, &deviceProperties);
        VkPhysicalDeviceFeatures deviceFeatures;
        vkGetPhysicalDeviceFeatures(device, &deviceFeatures);
        fprintf(stderr, "\tDevice: %s (%i, v%i) driver=%i\n", deviceProperties.deviceName, deviceProperties.deviceID, deviceProperties.apiVersion, deviceProperties.driverVersion);

        int score = 0;
        if (deviceProperties.deviceType == VK_PHYSICAL_DEVICE_TYPE_DISCRETE_GPU) {
            score += 1 << 10;
        }
        score += deviceProperties.limits.maxImageDimension2D;
        if (!deviceFeatures.geometryShader) {
            continue;
        }
        int firstGraphicsQueue = getFirstQueueFamilyOfType(device, VK_QUEUE_GRAPHICS_BIT);

        if (firstGraphicsQueue == -1) {
            continue;
        }
        VkBool32 presentSupport = VK_FALSE;
        vkGetPhysicalDeviceSurfaceSupportKHR(device, firstGraphicsQueue, surface, &presentSupport);
        if (presentSupport != VK_TRUE) {
            fprintf(stderr, "device doesn't support presentation along with graphics, skipping...\n");
            continue;
        }


        uint32_t deviceSurfaceFormatsCount;
        vkGetPhysicalDeviceSurfaceFormatsKHR(device, surface, &deviceSurfaceFormatsCount, NULL);
        if (deviceSurfaceFormatsCount == 0) {
            continue;
        }

        uint32_t devicePresentModeCount;
        vkGetPhysicalDeviceSurfacePresentModesKHR(device, surface, &devicePresentModeCount, NULL);
        if (devicePresentModeCount == 0) {
            continue;
        }

        if (score > best_score) {
            best_score = score;
            chosenDevice = device;
        }
    }
    free(devices);
    return chosenDevice;
}



void destroyDebug(VkInstance instance, VkDebugUtilsMessengerCreateInfoEXT debugCreateInfo, VkDebugUtilsMessengerEXT debugMessenger) {
    if (!debugCreateInfo.pfnUserCallback) {
        return;
    }
    PFN_vkDestroyDebugUtilsMessengerEXT handler = (PFN_vkDestroyDebugUtilsMessengerEXT) vkGetInstanceProcAddr(instance, "vkDestroyDebugUtilsMessengerEXT");
    if (handler) {
        handler(instance, debugMessenger, NULL);
    }
}

void recordCommandBuffer(VkCommandBuffer commandBuffer, uint32_t imageIndex) {
    VkCommandBufferBeginInfo commandBufferBeginInfo = {
        .sType = VK_STRUCTURE_TYPE_COMMAND_BUFFER_BEGIN_INFO,
        .flags = 0, // optional
        .pInheritanceInfo = NULL, // optional
    };

    if (vkBeginCommandBuffer(commandBuffer, &commandBufferBeginInfo) != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't begin command buffers\n");
        return;
    }
    VkClearValue clearColor = {{{0.0f, 0.0f, 0.0f, 1.0f}}};

    VkRenderPassBeginInfo renderPassInfo = {
        .sType = VK_STRUCTURE_TYPE_RENDER_PASS_BEGIN_INFO,
        .renderPass = renderPass,
        .framebuffer = swapChainFramebuffers[imageIndex],
        .renderArea = {
            .offset = {0, 0},
            .extent = swapChainExtent
        },
        .clearValueCount = 1,
        .pClearValues = &clearColor
    };
    vkCmdBeginRenderPass(commandBuffer, &renderPassInfo, VK_SUBPASS_CONTENTS_INLINE);
    vkCmdBindPipeline(commandBuffer, VK_PIPELINE_BIND_POINT_GRAPHICS, graphicsPipeline);
    VkViewport renderViewport = {
        .x = 0.0f,
        .y = 0.0f,
        .width = swapChainExtent.width,
        .height = swapChainExtent.height,
        .minDepth = 0.0f,
        .maxDepth = 1.0f,
    };
    vkCmdSetViewport(commandBuffer, 0, 1, &renderViewport);
    VkRect2D renderScissor = {
        .offset = {0, 0},
        .extent = swapChainExtent
    };
    vkCmdSetScissor(commandBuffer, 0, 1, &renderScissor);
    vkCmdDraw(
        commandBuffer,
        3, // how many vertexes -- triangle = 3
        1, // not doing instanced rendering, so 1
        0, // vertex offset
        0 // instance offset
    );
    vkCmdEndRenderPass(commandBuffer);
    if (vkEndCommandBuffer(commandBuffer) != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't run render pass\n");
    }
}

void drawFrame() {
    vkWaitForFences(device, 1, &inFlightFence, VK_TRUE, UINT64_MAX); // VK_TRUE == wait for all fences
    vkResetFences(device, 1, &inFlightFence);
    uint32_t imageIndex;
    vkAcquireNextImageKHR(device, swapChain, UINT64_MAX, imageAvailableSemaphore, VK_NULL_HANDLE, &imageIndex);
    vkResetCommandBuffer(commandBuffer, 0);
    recordCommandBuffer(commandBuffer, imageIndex);

    VkPipelineStageFlags waitStages[] = {VK_PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT};

    VkSemaphore signalSemaphores[] = {renderFinishedSemaphore};

    VkSubmitInfo submitInfo = {
        .sType = VK_STRUCTURE_TYPE_SUBMIT_INFO,
        .waitSemaphoreCount = 1,
        .pWaitSemaphores = &imageAvailableSemaphore,
        .pWaitDstStageMask = waitStages,
        .commandBufferCount = 1,
        .pCommandBuffers = &commandBuffer,
        .signalSemaphoreCount = 1,
        .pSignalSemaphores = signalSemaphores
    };

    if (vkQueueSubmit(graphicsQueue, 1, &submitInfo, inFlightFence) != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't submit to draw command buffer\n");
    }
    VkPresentInfoKHR presentInfo = {
        .sType = VK_STRUCTURE_TYPE_PRESENT_INFO_KHR,
        .waitSemaphoreCount = 1,
        .pWaitSemaphores = signalSemaphores,
        .swapchainCount = 1,
        .pSwapchains = &swapChain,
        .pImageIndices = &imageIndex,
        .pResults = NULL
    };
    vkQueuePresentKHR(presentQueue, &presentInfo);
}

int main(int argc, char* argv[]) {
    glfwSetErrorCallback(glfwErrorCallback);
    // init GLFW
    if (!glfwInit()) {
        fprintf(stderr, "glfw didn't initialize\n");
    }
    // init GLFW window
    glfwWindowHint(GLFW_CLIENT_API, GLFW_NO_API);
    GLFWwindow* window = glfwCreateWindow(800, 600, APP_NAME, NULL, NULL);
    if (!window) {
        fprintf(stderr, "glfw can't create window\n");
    }
    showExtensions();


    VkApplicationInfo appInfo = {
        .sType = VK_STRUCTURE_TYPE_APPLICATION_INFO,
        .pApplicationName = APP_NAME,
        .applicationVersion = VK_MAKE_VERSION(1, 0, 0),
        .pEngineName = "MWM D229-4 fundido no sol",
        .engineVersion = VK_MAKE_VERSION(1, 0, 0),
        .apiVersion = VK_API_VERSION_1_0
    };

    VkInstanceCreateInfo createInfo = {
        .sType = VK_STRUCTURE_TYPE_INSTANCE_CREATE_INFO,
        .pApplicationInfo = &appInfo,
        .enabledLayerCount = 0
    };
    getUsedExtensions(&createInfo);
    showValidationLayers();
    /* getUsedValidationLayers(&createInfo); */

    /* createInfo.enabledLayerCount = 1; */
    /* createInfo.ppEnabledLayerNames = VULKAN_VALIDATION_LAYERS; */

    if (vkCreateInstance(&createInfo, NULL, &instance) != VK_SUCCESS) {
        fprintf(stderr, "vulkan deu pau criando instância\n");
    }

    if (setupDebug(instance, &debugCreateInfo, &debugMessenger) != VK_SUCCESS) {
        fprintf(stderr, "falha ao dar setup no debug\n");
        debugCreateInfo.pfnUserCallback = NULL;
    }

    if (glfwCreateWindowSurface(instance, window, NULL, &surface) != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't create surface\n");
    }

    physicalDevice = getDevice(instance, surface);
    if (physicalDevice == VK_NULL_HANDLE) {
        fprintf(stderr, "falha ao achar um device compatível\n");
    }

    float queuePriority = 1.0f;
    VkDeviceQueueCreateInfo graphicsQueueCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_DEVICE_QUEUE_CREATE_INFO,
        .queueFamilyIndex = getFirstQueueFamilyOfType(physicalDevice, VK_QUEUE_GRAPHICS_BIT),
        .queueCount = 1,
        .pQueuePriorities = &queuePriority
    };
    VkDeviceQueueCreateInfo presentQueueCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_DEVICE_QUEUE_CREATE_INFO,
        .queueFamilyIndex = getFirstQueueFamilyOfType(physicalDevice, VK_QUEUE_GRAPHICS_BIT),
        .queueCount = 1,
        .pQueuePriorities = &queuePriority
    };

    VkDeviceQueueCreateInfo queueCreateInfos[] = {graphicsQueueCreateInfo, presentQueueCreateInfo};

    VkPhysicalDeviceFeatures deviceFeatures = {};
    VkDeviceCreateInfo deviceCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_DEVICE_CREATE_INFO,
        .pQueueCreateInfos = queueCreateInfos,
        .queueCreateInfoCount = 2,
        .pEnabledFeatures = &deviceFeatures,
        .enabledExtensionCount = 1,
        .ppEnabledExtensionNames = DEVICE_EXTENSIONS
    };
    // TODO: add validation layers here too, not required in newer implementations tho

    if (vkCreateDevice(physicalDevice, &deviceCreateInfo, NULL, &device) != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't create device\n");
    };

    // Maybe I can get some issues this part
    vkGetDeviceQueue(device, graphicsQueueCreateInfo.queueFamilyIndex, 0, &graphicsQueue);

    vkGetDeviceQueue(device, presentQueueCreateInfo.queueFamilyIndex, 0, &presentQueue);

    vkGetPhysicalDeviceSurfaceCapabilitiesKHR(physicalDevice, surface, &surfaceCapatibilitesDetails);

    swapChainExtent = surfaceCapatibilitesDetails.currentExtent;
    if (swapChainExtent.width == UINT32_MAX) {
        int windowWidth, windowHeight;
        glfwGetFramebufferSize(window, &windowWidth, &windowHeight);
        swapChainExtent.width = (uint32_t)windowWidth;
        swapChainExtent.height = (uint32_t)windowHeight;

        swapChainExtent.width = clamp(swapChainExtent.width, surfaceCapatibilitesDetails.minImageExtent.width, surfaceCapatibilitesDetails.maxImageExtent.width);
        swapChainExtent.height = clamp(swapChainExtent.height , surfaceCapatibilitesDetails.minImageExtent.height, surfaceCapatibilitesDetails.maxImageExtent.height);
    }
    uint32_t swapChainImageCount = clamp(surfaceCapatibilitesDetails.minImageCount + 1, surfaceCapatibilitesDetails.minImageCount, surfaceCapatibilitesDetails.maxImageCount);

    surfaceFormat = getSwapSurfaceFormat(physicalDevice, surface);
    presentMode = getSwapPresentMode(physicalDevice, surface);


    VkSwapchainCreateInfoKHR swapchainCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_SWAPCHAIN_CREATE_INFO_KHR,
        .surface = surface,
        .minImageCount = swapChainImageCount,
        .imageFormat = surfaceFormat.format,
        .imageColorSpace = surfaceFormat.colorSpace,
        .imageExtent = swapChainExtent,
        .imageUsage = VK_IMAGE_USAGE_COLOR_ATTACHMENT_BIT,
        // graphics queue == present queue so
        .imageSharingMode = VK_SHARING_MODE_EXCLUSIVE,
        .queueFamilyIndexCount = 0,
        .pQueueFamilyIndices = NULL,
        .preTransform = surfaceCapatibilitesDetails.currentTransform,
        .compositeAlpha = VK_COMPOSITE_ALPHA_OPAQUE_BIT_KHR, // do not blend with other windows
        .presentMode = presentMode,
        .clipped = VK_TRUE, // render pixels that are not shown?
        .oldSwapchain = VK_NULL_HANDLE
    };

    if (vkCreateSwapchainKHR(device, &swapchainCreateInfo, NULL, &swapChain) != VK_SUCCESS) {
        fprintf(stderr, "can't create swap chain\n");
    };

    uint32_t swapchainImageCount;
    vkGetSwapchainImagesKHR(device, swapChain, &swapChainImageCount, NULL);
    swapchainImages = malloc(sizeof(VkImage)*swapChainImageCount);
    swapchainImageViews = malloc(sizeof(VkImageView)*swapChainImageCount);
    vkGetSwapchainImagesKHR(device, swapChain, &swapChainImageCount, swapchainImages);
    swapChainImageFormat = surfaceFormat.format;

    for (int i = 0; i < swapchainImageCount; i++) {
        VkImageViewCreateInfo imageViewCreateInfo = {
            .sType = VK_STRUCTURE_TYPE_IMAGE_VIEW_CREATE_INFO,
            .image = swapchainImages[i],
            .viewType = VK_IMAGE_VIEW_TYPE_2D,
            .format = swapChainImageFormat,
            .components = {
                .r = VK_COMPONENT_SWIZZLE_IDENTITY,
                .g = VK_COMPONENT_SWIZZLE_IDENTITY,
                .b = VK_COMPONENT_SWIZZLE_IDENTITY,
                .a = VK_COMPONENT_SWIZZLE_IDENTITY,
            },
            .subresourceRange = {
                .aspectMask = VK_IMAGE_ASPECT_COLOR_BIT,
                .baseMipLevel = 0,
                .levelCount = 1,
                .baseArrayLayer = 0,
                .layerCount = 1
            }
        };
        if (vkCreateImageView(device, &imageViewCreateInfo, NULL, &swapchainImageViews[i]) != VK_SUCCESS) {
            fprintf(stderr, "vulkan: can't create image view\n");
        }
    }

    if (readShader(device, &vertexShaderModule, "./vert.spv") != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't create vertex shader module\n");
    }

    if (readShader(device, &fragmentShaderModule, "./frag.spv") != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't create fragment shader module\n");
    }

    VkPipelineShaderStageCreateInfo vertexShaderStageCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO,
        .stage = VK_SHADER_STAGE_VERTEX_BIT,
        .module = vertexShaderModule,
        .pName = "main"
    };

    VkPipelineShaderStageCreateInfo fragmentShaderStageCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO,
        .stage = VK_SHADER_STAGE_FRAGMENT_BIT,
        .module = fragmentShaderModule,
        .pName = "main"
    };

    VkPipelineShaderStageCreateInfo shaderStages[] = {vertexShaderStageCreateInfo, fragmentShaderStageCreateInfo};

    VkDynamicState dynamicStates[] = {
        VK_DYNAMIC_STATE_VIEWPORT,
        VK_DYNAMIC_STATE_SCISSOR,
    };

    VkPipelineDynamicStateCreateInfo dynamicStateCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_PIPELINE_DYNAMIC_STATE_CREATE_INFO,
        .dynamicStateCount = 2,
        .pDynamicStates = dynamicStates
    };

    VkPipelineVertexInputStateCreateInfo vertexInputStateCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_PIPELINE_VERTEX_INPUT_STATE_CREATE_INFO,
        .vertexBindingDescriptionCount = 0,
        .pVertexBindingDescriptions = NULL, // optional
        .vertexAttributeDescriptionCount = 0,
        .pVertexAttributeDescriptions = NULL, // optional
    };

    VkPipelineInputAssemblyStateCreateInfo inputAssemblyCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_PIPELINE_INPUT_ASSEMBLY_STATE_CREATE_INFO,
        .topology = VK_PRIMITIVE_TOPOLOGY_TRIANGLE_LIST,
        .primitiveRestartEnable = VK_FALSE
    };

    VkViewport viewport = {
        .x = 0.0f,
        .y = 0.0f,
        .width = (float) swapChainExtent.width,
        .height = (float) swapChainExtent.height,
        .minDepth = 0.0f,
        .maxDepth = 1.0f
    };

    VkOffset2D scissorOffset = {0, 0};
    VkRect2D scissor = {
        .offset = scissorOffset,
        .extent = swapChainExtent
    };

    VkPipelineViewportStateCreateInfo viewportStateCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_PIPELINE_VIEWPORT_STATE_CREATE_INFO,
        .viewportCount = 1,
        .pViewports = &viewport,
        .scissorCount = 1,
        .pScissors = &scissor
    };

    VkPipelineRasterizationStateCreateInfo rasterizationStateCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_PIPELINE_RASTERIZATION_STATE_CREATE_INFO,
        .depthClampEnable = VK_FALSE,
        .rasterizerDiscardEnable = VK_FALSE,
        .polygonMode = VK_POLYGON_MODE_FILL,
        .lineWidth = 1.0f,
        .cullMode = VK_CULL_MODE_BACK_BIT,
        .frontFace = VK_FRONT_FACE_CLOCKWISE,
        .depthBiasEnable = VK_FALSE,
        .depthBiasConstantFactor = 0.0f, // optional
        .depthBiasClamp = 0.0f, // optional
        .depthBiasSlopeFactor = 0.0f, // optional
    };

    VkPipelineMultisampleStateCreateInfo multisampleStateCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_PIPELINE_MULTISAMPLE_STATE_CREATE_INFO,
        .sampleShadingEnable = VK_FALSE,
        .rasterizationSamples = VK_SAMPLE_COUNT_1_BIT,
        .minSampleShading = 1.0f, // optional
        .pSampleMask = NULL, // optional
        .alphaToCoverageEnable = VK_FALSE, // optional
        .alphaToOneEnable = VK_FALSE // optional
    };

    VkPipelineColorBlendAttachmentState colorBlendAttachment = {
        .colorWriteMask =
              VK_COLOR_COMPONENT_R_BIT
            | VK_COLOR_COMPONENT_G_BIT
            | VK_COLOR_COMPONENT_B_BIT
            | VK_COLOR_COMPONENT_A_BIT,
        .blendEnable = VK_FALSE,
        .srcColorBlendFactor = VK_BLEND_FACTOR_ONE, // optional
        .dstColorBlendFactor = VK_BLEND_FACTOR_ZERO, // optional
        .colorBlendOp = VK_BLEND_OP_ADD, // optional
        .srcAlphaBlendFactor = VK_BLEND_FACTOR_ONE, // optional
        .dstAlphaBlendFactor = VK_BLEND_FACTOR_ZERO, // optional
        .alphaBlendOp = VK_BLEND_OP_ADD // optional
    };

    VkPipelineColorBlendStateCreateInfo colorBlendStateCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_PIPELINE_COLOR_BLEND_STATE_CREATE_INFO,
        .logicOpEnable = VK_FALSE,
        .logicOp = VK_LOGIC_OP_COPY,
        .attachmentCount = 1,
        .pAttachments = &colorBlendAttachment,
        .blendConstants = { 0.0f, 0.0f, 0.0f, 0.0f }
    };

    VkPipelineLayoutCreateInfo pipelineLayoutCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_PIPELINE_LAYOUT_CREATE_INFO,
        .setLayoutCount = 0, // optional
        .pSetLayouts = NULL, // optional
        .pushConstantRangeCount = 0, // optional
        .pPushConstantRanges = NULL
    };

    if (vkCreatePipelineLayout(device, &pipelineLayoutCreateInfo, NULL, &pipelineLayout) != VK_SUCCESS) {
        fprintf(stderr, "vulkan: não foi possivel criar o pipeline layout\n");
    }

    VkAttachmentDescription colorAttachment = {
        .format = swapChainImageFormat,
        .samples = VK_SAMPLE_COUNT_1_BIT,
        .loadOp = VK_ATTACHMENT_LOAD_OP_CLEAR,
        .storeOp = VK_ATTACHMENT_STORE_OP_STORE,
        .stencilLoadOp = VK_ATTACHMENT_LOAD_OP_DONT_CARE,
        .stencilStoreOp = VK_ATTACHMENT_STORE_OP_DONT_CARE,
        .initialLayout = VK_IMAGE_LAYOUT_UNDEFINED,
        .finalLayout = VK_IMAGE_LAYOUT_PRESENT_SRC_KHR
    };

    VkAttachmentReference colorAttachmentRef = {
        .attachment = 0,
        .layout = VK_IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL
    };

    // this specify about where to get the color from the shader
    VkSubpassDescription subpassDescription = {
        .pipelineBindPoint = VK_PIPELINE_BIND_POINT_GRAPHICS,
        .colorAttachmentCount = 1,
        .pColorAttachments = &colorAttachmentRef
    };

    VkSubpassDependency subpassDependency = {
        .srcSubpass = VK_SUBPASS_EXTERNAL,
        .dstSubpass = 0,
        .srcStageMask = VK_PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT,
        .srcAccessMask = 0,
        .dstStageMask = VK_PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT,
        .dstAccessMask = VK_ACCESS_COLOR_ATTACHMENT_WRITE_BIT
    };

    VkRenderPassCreateInfo renderPassCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_RENDER_PASS_CREATE_INFO,
        .attachmentCount = 1,
        .pAttachments = &colorAttachment,
        .subpassCount = 1,
        .pSubpasses = &subpassDescription,
        .dependencyCount = 1,
        .pDependencies = &subpassDependency
    };

    if (vkCreateRenderPass(device, &renderPassCreateInfo, NULL, &renderPass) != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't create render pass\n");
    }

    VkGraphicsPipelineCreateInfo graphicsPipelineCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_GRAPHICS_PIPELINE_CREATE_INFO,
        .stageCount = 2,
        .pStages = shaderStages,
        .pVertexInputState = &vertexInputStateCreateInfo,
        .pInputAssemblyState = &inputAssemblyCreateInfo,
        .pViewportState = &viewportStateCreateInfo,
        .pRasterizationState = &rasterizationStateCreateInfo,
        .pMultisampleState = &multisampleStateCreateInfo,
        .pDepthStencilState = NULL, // optional
        .pColorBlendState = &colorBlendStateCreateInfo,
        .pDynamicState = &dynamicStateCreateInfo,
        .layout = pipelineLayout,
        .renderPass = renderPass,
        .subpass = 0,
        .basePipelineHandle = VK_NULL_HANDLE,
        .basePipelineIndex = -1 // optional
    };


    if (vkCreateGraphicsPipelines(device, VK_NULL_HANDLE, 1, &graphicsPipelineCreateInfo, NULL, &graphicsPipeline) != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't create graphics pipeline\n");
    }

    swapChainFramebuffers = malloc(sizeof(VkFramebuffer)*swapChainImageCount);
    for (int i = 0; i < swapChainImageCount; i++) {
        VkImageView attachments[] = {
            swapchainImageViews[i]
        };
        VkFramebufferCreateInfo framebufferCreateInfo = {
            .sType = VK_STRUCTURE_TYPE_FRAMEBUFFER_CREATE_INFO,
            .renderPass = renderPass,
            .attachmentCount = 1,
            .pAttachments = attachments,
            .width = swapChainExtent.width,
            .height = swapChainExtent.height,
            .layers = 1
        };
        if (vkCreateFramebuffer(device, &framebufferCreateInfo, NULL, &swapChainFramebuffers[i]) != VK_SUCCESS) {
            fprintf(stderr, "vulkan: can't create framebuffer\n");
        }
    }

    VkCommandPoolCreateInfo commandPoolCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_COMMAND_POOL_CREATE_INFO,
        .flags = VK_COMMAND_POOL_CREATE_RESET_COMMAND_BUFFER_BIT,
        .queueFamilyIndex = graphicsQueueCreateInfo.queueFamilyIndex
    };
    if (vkCreateCommandPool(device, &commandPoolCreateInfo, NULL, &commandPool) != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't create command pool\n");
    }

    VkCommandBufferAllocateInfo commandBufferAllocateInfo = {
        .sType = VK_STRUCTURE_TYPE_COMMAND_BUFFER_ALLOCATE_INFO,
        .commandPool = commandPool,
        .level = VK_COMMAND_BUFFER_LEVEL_PRIMARY,
        .commandBufferCount = 1
    };
    if (vkAllocateCommandBuffers(device, &commandBufferAllocateInfo, &commandBuffer) != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't allocate command buffers\n");
    }

    VkSemaphoreCreateInfo semaphoreCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_SEMAPHORE_CREATE_INFO
    };

    VkFenceCreateInfo fenceCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_FENCE_CREATE_INFO,
        .flags =
            VK_FENCE_CREATE_SIGNALED_BIT // doesn't block on the first pass
    };

    if (vkCreateSemaphore(device, &semaphoreCreateInfo, NULL, &imageAvailableSemaphore) != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't create imageAvailableSemaphore\n");
    }
    if (vkCreateSemaphore(device, &semaphoreCreateInfo, NULL, &renderFinishedSemaphore) != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't create renderFinishedSemaphore\n");
    }
    if (vkCreateFence(device, &fenceCreateInfo, NULL, &inFlightFence) != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't create inFlightFence\n");
    }

    // Paused at: https://vulkan-tutorial.com/en/Drawing_a_triangle/Graphics_pipeline_basics/Conclusion


    fprintf(stderr, "Chegou agui\n");
    while(!glfwWindowShouldClose(window)) {
        glfwPollEvents();
        drawFrame();
    }

    fprintf(stderr, "E agui\n");
    vkDestroySemaphore(device, imageAvailableSemaphore, NULL);
    vkDestroySemaphore(device, renderFinishedSemaphore, NULL);
    vkDestroyFence(device, inFlightFence, NULL);

    vkDestroyCommandPool(device, commandPool, NULL);

    for (int i = 0; i < swapchainImageCount; i++) {
        vkDestroyFramebuffer(device, swapChainFramebuffers[i], NULL);
    }
    free(swapChainFramebuffers);

    vkDestroyPipeline(device, graphicsPipeline, NULL);
    vkDestroyPipelineLayout(device, pipelineLayout, NULL);
    vkDestroyRenderPass(device, renderPass, NULL);

    vkDestroyPipelineLayout(device, pipelineLayout, NULL);

    vkDestroyShaderModule(device, vertexShaderModule, NULL);
    vkDestroyShaderModule(device, fragmentShaderModule, NULL);

    for (int i = 0; i < swapchainImageCount; i++) {
        vkDestroyImageView(device, swapchainImageViews[i], NULL);
    }
    free(swapchainImageViews);
    free(swapchainImages);

    vkDestroySwapchainKHR(device, swapChain, NULL);
    vkDestroySurfaceKHR(instance, surface, NULL);
    vkDestroyInstance(instance, NULL);
    vkDestroyDevice(device, NULL);

    // deinit GLFW window
    glfwDestroyWindow(window);

    /* destroyDebug(instance, debugCreateInfo, debugMessenger); */
    // deinit GLFW
    glfwTerminate();
    return 0;
}
