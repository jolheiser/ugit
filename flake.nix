{
  description = "Minimal git server";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    tailwind-ctp = {
      url = "git+https://git.jojodev.com/jolheiser/tailwind-ctp";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    tailwind-ctp-lsp = {
      url = "git+https://git.jojodev.com/jolheiser/tailwind-ctp-intellisense";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    self,
    nixpkgs,
    tailwind-ctp,
    tailwind-ctp-lsp,
  } @ inputs: let
    system = "x86_64-linux";
    pkgs = nixpkgs.legacyPackages.${system};
    tailwind-ctp = inputs.tailwind-ctp.packages.${system}.default;
    tailwind-ctp-lsp = inputs.tailwind-ctp-lsp.packages.${system}.default;
    ugit = pkgs.buildGoModule rec {
      pname = "ugitd";
      version = "0.0.1";
      src = pkgs.nix-gitignore.gitignoreSource [] (builtins.path {
        name = pname;
        path = ./.;
      });
      subPackages = ["cmd/ugitd"];
      vendorHash = "sha256-8kI94hcJupAUye6cEAmIlN+CrtYSXlgoAlmpyXArfF8=";
      meta = with pkgs.lib; {
        description = "Minimal git server";
        homepage = "https://git.jolheiser.com/ugit";
        maintainers = with maintainers; [jolheiser];
        mainProgram = "ugitd";
      };
    };
  in {
    packages.${system}.default = ugit;
    devShells.${system}.default = pkgs.mkShell {
      nativeBuildInputs = with pkgs; [
        go
        gopls
        templ
        tailwind-ctp
        tailwind-ctp-lsp
        vscode-langservers-extracted
      ];
    };
    nixosModules.default = {
      pkgs,
      lib,
      config,
      ...
    }: let
      cfg = config.services.ugit;
      yamlFormat = pkgs.formats.yaml {};
      configFile = pkgs.writeText "ugit.yaml" (builtins.readFile (yamlFormat.generate "ugit-yaml" cfg.config));
      authorizedKeysFile = pkgs.writeText "ugit_keys" (builtins.concatStringsSep "\n" cfg.authorizedKeys);
    in {
      options = with lib; {
        services.ugit = {
          enable = mkEnableOption "Enable ugit";

          package = mkOption {
            type = types.package;
            description = "ugit package to use";
            default = ugit;
          };

          repoDir = mkOption {
            type = types.str;
            description = "where ugit stores repositories";
            default = "/var/lib/ugit/repos";
          };

          authorizedKeys = mkOption {
            type = types.listOf types.str;
            description = "list of keys to use for authorized_keys";
            default = [];
          };

          authorizedKeysFile = mkOption {
            type = types.str;
            description = "path to authorized_keys file ugit uses for auth";
            default = "/var/lib/ugit/authorized_keys";
          };

          hostKeyFile = mkOption {
            type = types.str;
            description = "path to host key file (will be created if it doesn't exist)";
            default = "/var/lib/ugit/ugit_ed25519";
          };

          config = mkOption {
            type = types.attrs;
            default = {};
            description = "config.yaml contents";
          };

          user = mkOption {
            type = types.str;
            default = "ugit";
            description = "User account under which ugit runs";
          };

          group = mkOption {
            type = types.str;
            default = "ugit";
            description = "Group account under which ugit runs";
          };

          debug = mkOption {
            type = types.bool;
            default = false;
          };

          openFirewall = mkOption {
            type = types.bool;
            default = false;
          };
        };
      };
      config = lib.mkIf cfg.enable {
        users.users."${cfg.user}" = {
          home = "/var/lib/ugit";
          createHome = true;
          group = "${cfg.group}";
          isSystemUser = true;
          isNormalUser = false;
          description = "user for ugit service";
        };
        users.groups."${cfg.group}" = {};
        networking.firewall = lib.mkIf cfg.openFirewall {
          allowedTCPPorts = [8448 8449];
        };

        systemd.services.ugit = {
          enable = true;
          script = let
            authorizedKeysPath =
              if (builtins.length cfg.authorizedKeys) > 0
              then authorizedKeysFile
              else cfg.authorizedKeysFile;
            args = ["--config=${configFile}" "--repo-dir=${cfg.repoDir}" "--ssh.authorized-keys=${authorizedKeysPath}" "--ssh.host-key=${cfg.hostKeyFile}"] ++ lib.optionals cfg.debug ["--debug"];
          in "${cfg.package}/bin/ugitd ${builtins.concatStringsSep " " args}";
          wantedBy = ["multi-user.target"];
          after = ["network-online.target"];
          serviceConfig = {
            User = cfg.user;
            Group = cfg.group;
            Restart = "always";
            RestartSec = "15";
            WorkingDirectory = "/var/lib/ugit";
          };
        };
      };
    };
  };
}
