#version 450

layout(location = 0) out vec4 outColor;
layout(location = 0) in vec3 fragColor;

void main() {
    // outColor = vec4(1.0, 0.0, 0.0, 1.0);
    outColor = vec4(fragColor, 1.0);
}
