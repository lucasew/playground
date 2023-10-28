#include<stdio.h>
#include<stdlib.h>
#include<string.h>

#define GLFW_INCLUDE_VULKAN

#include<GLFW/glfw3.h>

const char* APP_NAME = "V U L K A N";
const char* const VULKAN_VALIDATION_LAYERS[] = {
    "VK_LAYER_KHRONOS_validation"
};

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

    // Paused at: https://vulkan-tutorial.com/en/Drawing_a_triangle/Setup/Validation_layers


    fprintf(stderr, "Chegou agui\n");
    while(!glfwWindowShouldClose(window)) {
        glfwPollEvents();
    }

    // deinit GLFW window
    glfwDestroyWindow(window);

    destroyDebug(instance, debugCreateInfo, debugMessenger);
    // deinit GLFW
    glfwTerminate();
    return 0;
}
