{ pkgs ? import <nixpkgs> {}}:

let
  image = pkgs.dockerTools.pullImage {
    imageName = "hello-world";
    sha256 = "sha256-pi33xlJgmjrPI9CqmwG1FW6mXN9tuUh69JT1hjH+uRc=";
    imageDigest = "sha256:e2fc4e5012d16e7fe466f5291c476431beaa1f9b90a5c2125b493ed28e2aba57";
  };

  extracted = pkgs.runCommand "container-extracted" {
    buildInputs = with pkgs; [ jq makeWrapper which ];
  } ''
    mkdir -p $out/image-data $out/bin
    mkdir -p extracted-image
    tar -xvf ${image} -C extracted-image
    cp -r extracted-image $out/extracted-image
    cat extracted-image/manifest.json | jq

    CONFIG_FILE="$(cat extracted-image/manifest.json | jq -r '.[].Config')"

    cat extracted-image/$CONFIG_FILE | jq

    cat extracted-image/$CONFIG_FILE | jq '.rootfs.diff_ids | .[]'  -r \
      | while IFS=":" read -r type hash; do
        tar -xvf extracted-image/$hash.tar -C $out/image-data
      done

    WorkingDirectory="$(cat extracted-image/$CONFIG_FILE | jq '.config.WorkingDir'  -r)"

    makeWrapperArgs=()
    makeWrapperArgs+=(--add-flags -p)
    makeWrapperArgs+=(--add-flags "WorkingDirectory=$WorkingDirectory")

    while IFS== read -r key value ; do
      echo env $key $value
      makeWrapperArgs+=(--add-flags "-E")
      makeWrapperArgs+=(--add-flags "$key=$value")
    done < <(cat extracted-image/$CONFIG_FILE | jq -r '.config.Env | .[]')

    echo prefix $prefixCommand

    makeWrapperArgs+=(--prefix PATH : ${pkgs.lib.makeBinPath [ pkgs.coreutils pkgs.xorg.lndir ]})
    # makeWrapperArgs+=(--add-flags -p)
    # makeWrapperArgs+=(--add-flags TemporaryFileSystem=/:ro)

    # the extracted image folder is readonly so making a writable overlay with symlinks pointing to the nix store
    makeWrapperArgs+=(--run 'rootfs=$(mktemp -d)')
    makeWrapperArgs+=(--run "rootfs_base=$out/image-data")
    makeWrapperArgs+=(--run 'lndir $rootfs_base $rootfs')
    makeWrapperArgs+=(--add-flags -p)
    makeWrapperArgs+=(--add-flags 'RootDirectory=$rootfs')

    # nix store must be available for the symlinks to work
    makeWrapperArgs+=(--add-flags -p)
    makeWrapperArgs+=(--add-flags 'BindReadOnlyPaths=/nix/store')

    # debugging
    makeWrapperArgs+=(--run 'echo rootfs $rootfs >&2')
    makeWrapperArgs+=(--run 'echo args $@ >&2')
    
    echo ${"$"}{makeWrapperArgs[@]}
    makeWrapper ${pkgs.systemd}/bin/systemd-run $out/bin/run-container "${"$"}{makeWrapperArgs[@]}"
    # makeWrapper $(which echo) $out/bin/run-container "${"$"}{makeWrapperArgs[@]}"

  '';
in extracted
