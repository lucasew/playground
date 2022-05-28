{ lib
, ...
}:
let
  inherit (lib) types mkOption;
  inherit (types) str strMatching attrsOf listOf submodule nullOr nonEmptyListOf;
  flag = submodule ({config, ...}: {
    options = {
      keywords = mkOption {
        type = nonEmptyListOf (strMatching "-[a-zA-Z0-9]|-(-[a-z0-9]*)");
        default = [];
        description = "Which keywords refer to this flag";
      };
      description = mkOption {
        type = str;
        default = "";
        description = "Description of the flag value";
      };
      validator = mkOption {
        type = str;
        default = "any";
        description = "Command to run passing the input to validate the flag value";
      };
      variable = mkOption {
        type = strMatching "[A-Z][A-Z_]*";
        description = "Variable to store the result";
      };
    };
  });
  command = {
    name = mkOption {
      type = strMatching "[a-z][a-z0-9_\\-]*";
      default = "example";
      description = "Name of the command shown on --help";
    };
    description = mkOption {
      type = str;
      default = "Example cli script generated with nix";
      description = "Command description";
    };
    flags = mkOption {
      type = listOf flag;
      default = [];
      description = "Command flags";
    };
    subcommands = mkOption {
      type = nullOr (attrsOf (submodule ({...}: {options = command; })));
      default = {};
      description = "Subcommands";
    };
    action = mkOption {
      type = str;
      default = "exit 0";
      description = "Action itself of the command or subcommand";
    };
  };
in {
  options = command;
}
