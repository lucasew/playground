#include<stdio.h>
#include<stdlib.h>
#include<string.h>
#include<limits.h>
#include<stdint.h>

#define GLFW_INCLUDE_VULKAN

#include<GLFW/glfw3.h>

const char* APP_NAME = "V U L K A N";
const char* const VULKAN_VALIDATION_LAYERS[] = {
    "VK_LAYER_KHRONOS_validation"
};

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


    VkInstance instance;
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

    VkDebugUtilsMessengerCreateInfoEXT debugCreateInfo;
    VkDebugUtilsMessengerEXT debugMessenger;
    if (setupDebug(instance, &debugCreateInfo, &debugMessenger) != VK_SUCCESS) {
        fprintf(stderr, "falha ao dar setup no debug\n");
        debugCreateInfo.pfnUserCallback = NULL;
    }

    VkSurfaceKHR surface;
    if (glfwCreateWindowSurface(instance, window, NULL, &surface) != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't create surface\n");
    }

    VkPhysicalDevice physicalDevice = getDevice(instance, surface);
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
    const char * const DEVICE_EXTENSIONS[] = {
        "VK_KHR_swapchain"
    };
    VkDeviceCreateInfo deviceCreateInfo = {
        .sType = VK_STRUCTURE_TYPE_DEVICE_CREATE_INFO,
        .pQueueCreateInfos = queueCreateInfos,
        .queueCreateInfoCount = 2,
        .pEnabledFeatures = &deviceFeatures,
        .enabledExtensionCount = 1,
        .ppEnabledExtensionNames = DEVICE_EXTENSIONS
    };
    // TODO: add validation layers here too, not required in newer implementations tho

    VkDevice device;
    if (vkCreateDevice(physicalDevice, &deviceCreateInfo, NULL, &device) != VK_SUCCESS) {
        fprintf(stderr, "vulkan: can't create device\n");
    };

    // Maybe I can get some issues this part
    VkQueue graphicsQueue;

    vkGetDeviceQueue(device, graphicsQueueCreateInfo.queueFamilyIndex, 0, &graphicsQueue);

    VkQueue presentQueue;
    vkGetDeviceQueue(device, presentQueueCreateInfo.queueFamilyIndex, 0, &presentQueue);

    VkSurfaceCapabilitiesKHR surfaceCapatibilitesDetails;
    vkGetPhysicalDeviceSurfaceCapabilitiesKHR(physicalDevice, surface, &surfaceCapatibilitesDetails);

    VkExtent2D swapChainExtent = surfaceCapatibilitesDetails.currentExtent;
    if (swapChainExtent.width == UINT32_MAX) {
        int windowWidth, windowHeight;
        glfwGetFramebufferSize(window, &windowWidth, &windowHeight);
        swapChainExtent.width = (uint32_t)windowWidth;
        swapChainExtent.height = (uint32_t)windowHeight;

        swapChainExtent.width = clamp(swapChainExtent.width, surfaceCapatibilitesDetails.minImageExtent.width, surfaceCapatibilitesDetails.maxImageExtent.width);
        swapChainExtent.height = clamp(swapChainExtent.height , surfaceCapatibilitesDetails.minImageExtent.height, surfaceCapatibilitesDetails.maxImageExtent.height);
    }
    uint32_t swapChainImageCount = clamp(surfaceCapatibilitesDetails.minImageCount + 1, surfaceCapatibilitesDetails.minImageCount, surfaceCapatibilitesDetails.maxImageCount);

    VkSurfaceFormatKHR surfaceFormat = getSwapSurfaceFormat(physicalDevice, surface);
    VkPresentModeKHR presentMode = getSwapPresentMode(physicalDevice, surface);


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

    VkSwapchainKHR swapChain;
    if (vkCreateSwapchainKHR(device, &swapchainCreateInfo, NULL, &swapChain) != VK_SUCCESS) {
        fprintf(stderr, "can't create swap chain\n");
    };

    uint32_t swapchainImageCount;
    vkGetSwapchainImagesKHR(device, swapChain, &swapChainImageCount, NULL);
    VkImage* swapchainImages = malloc(sizeof(VkImage)*swapChainImageCount);
    VkImageView* swapchainImageViews = malloc(sizeof(VkImageView)*swapChainImageCount);
    vkGetSwapchainImagesKHR(device, swapChain, &swapChainImageCount, swapchainImages);
    VkFormat swapChainImageFormat = surfaceFormat.format;

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

    // Paused at: https://vulkan-tutorial.com/en/Drawing_a_triangle/Graphics_pipeline_basics/Introduction
    fprintf(stderr, "Chegou agui\n");
    while(!glfwWindowShouldClose(window)) {
        glfwPollEvents();
    }
    fprintf(stderr, "E agui\n");

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