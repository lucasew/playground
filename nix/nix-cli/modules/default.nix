{lib, config, ...}:
let
  inherit (lib) mkOption types optionalString;
  inherit (builtins) concatStringsSep length attrValues mapAttrs;
in
{
  imports = [
    ./base.nix
    ./validator.nix
  ];
  options = {
    target = mkOption {
      type = types.attrsOf types.str;
      description = "Outputs";
    };
    shebang = mkOption {
      type = types.str;
      description = "Script shebang";
      default = "#!/usr/bin/env bash";
    };
  };
  config = {
    target.shellscript = let
      buildHelp = cfg: let
          subcommandTree = cfg._subcommand;
          subcommandTree' = concatStringsSep " " subcommandTree;

          firstLine = "$(bold '${subcommandTree'}') ${cfg.description}";
          firstLine' = "echo \"${firstLine}\"";
          subcommands = mapAttrs (k: v: v.description) cfg.subcommands;
          subcommands' = mapAttrs (k: v: ''
            printf "\t"
            printf "$(bold '${k}'): "
            echo '${v}'
          '') subcommands;
          subcommands'' = attrValues subcommands';
          subcommands''' = concatStringsSep "\n" subcommands'';

          hasSubcommands = length (attrValues subcommands) > 0;
          
          flag2txt = flag: let
            keywordLine = flag.keywords;
            keywordLine' = concatStringsSep ", " keywordLine;
          in ''
            printf "\t"
            echo "$(bold '${keywordLine'}, ${flag.variable}') (${flag.validator}): ${flag.description}"
          '';

          flags = cfg.flags;
          flags' = [''
            printf "\t"
            echo "$(bold '-h, --help'): Show this help message"
          ''] ++ (map flag2txt flags);
          flags'' = concatStringsSep "\n" flags';
        in ''
          ${firstLine'}
          ${optionalString hasSubcommands ''printf "\nSubcommands\n"''}
          ${subcommands'''}
          printf "\nFlags\n"
          ${flags''}
          exit 0
        '';
      buildCommandTree = cfg: let
        help = buildHelp cfg;

        mkSubcommandHandler = k: v: let
          flags = cfg.flags ++ v.flags;
          _subcommand = cfg._subcommand ++ [ v.name ];
          subcommandArgs = v // { inherit flags _subcommand; };
          handler = buildCommandTree subcommandArgs;
          handler' = ''
            ${k})
              shift
              ${handler}
              exit 0
            ;;
          '';
          in handler';

        mkFlagHandler = flag:
        let
          caseExpr = concatStringsSep " | " flag.keywords;
          varAttrExpr = ''${flag.variable}="$1"'';
          validateExprError = ''echo "flag '$flagkey' (${flag.variable}) doesn't pass the validation as a ${flag.validator}" '';

          isBool = flag.validator == "bool";
          validateExpr = optionalString (!isBool) ''validate_${flag.validator} "$arg" || ${validateExprError}'';
        in ''
          ${caseExpr} )
            shift
            ${optionalString isBool "export ${flag.variable}=1"}
            ${optionalString isBool "continue"}

            if [ $# -eq 0 ]; then
              error "the flag '$flagkey' expects a value of type ${flag.validator} but found end of parameters"
            fi

            arg="$1"; shift

            export ${flag.variable}="$arg"
            ${validateExpr}
            break
          ;;
        '';

        subcommands = cfg.subcommands;
        subcommands' = mapAttrs mkSubcommandHandler subcommands;
        subcommands'' = attrValues subcommands';
        subcommands''' = concatStringsSep "\n" subcommands'';
        flags = cfg.flags;
        flags' = map mkFlagHandler flags;
        flags'' = concatStringsSep "\n" flags';
      in ''
        if [ $# -eq 0 ]; then
          ${help}
        fi
        case "$1" in
          ${subcommands'''}
        esac
        ARGS=()
        while [ ! $# -eq 0 ]; do
          local flagkey="$1"

          case "$flagkey" in
              -h | --help)
                ${help}
              ;;
                ${flags''}
              *)
                error "invalid keyword argument near '$flagkey'"
              ;;
          esac
        done
        ${cfg.action}
        exit 0
        '';
    in ''
        ${config.shebang}
        set -eu
        function error {
          echo "error: $@" >&2
          exit 1
        }
        function bold {
          if which tput >/dev/null 2>/dev/null; then
            printf "$(tput bold)$*$(tput sgr0)"
          else
            printf "$*"
          fi
        }
        ${builtins.concatStringsSep "\n" (builtins.attrValues (builtins.mapAttrs (k: v: ''
          function validate_${k} {
            ${v}
          }
        '') config.validators))}

          function _main() {
            ${buildCommandTree (config // {
              _subcommand = [ config.name ];
            })}
          }
          _main "$@"
    '';
  };
}
