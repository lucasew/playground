{ pkgs ? import <nixpkgs> {} }:

{
  objs = pkgs.callPackage ./example { };
}
