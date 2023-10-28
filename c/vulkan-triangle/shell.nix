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
}
