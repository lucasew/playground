{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    gnumake
    vulkan-headers
    vulkan-loader
    vulkan-validation-layers
    glfw
    ccls
    xorg.libX11
    xorg.libXrandr
    xorg.libXi
    xorg.libXxf86vm
  ];

  VULKAN_SDK = "${pkgs.vulkan-validation-layers}/share/vulkan/explicit_layer.d";
}
