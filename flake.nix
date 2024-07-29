{
  description = "Minimal git server";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    tailwind-ctp = {
      url = "git+https://git.jolheiser.com/tailwind-ctp";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    tailwind-ctp-lsp = {
      url = "git+https://git.jolheiser.com/tailwind-ctp-intellisense";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    self,
    nixpkgs,
    gomod2nix,
    tailwind-ctp,
    tailwind-ctp-lsp,
  } @ inputs: let
    system = "x86_64-linux";
    pkgs = nixpkgs.legacyPackages.${system};
    tailwind-ctp = inputs.tailwind-ctp.packages.${system}.default;
    tailwind-ctp-lsp = inputs.tailwind-ctp-lsp.packages.${system}.default;
    ugit = gomod2nix.legacyPackages.${system}.buildGoApplication rec {
      name = "ugitd";
      src = pkgs.nix-gitignore.gitignoreSource [] (builtins.path {
        inherit name;
        path = ./.;
      });
      pwd = ./.;
      subPackages = ["cmd/ugitd" "cmd/ugit-uci"];
      CGO_ENABLED = 0;
      flags = [
        "-trimpath"
      ];
      ldflags = [
        "-s"
        "-w"
        "-extldflags -static"
      ];
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
        gomod2nix.legacyPackages.${system}.gomod2nix
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
      options = let
        inherit (lib) mkEnableOption mkOption types;
      in {
        services.ugit = {
          enable = mkEnableOption "Enable ugit";

          package = mkOption {
            type = types.package;
            description = "ugit package to use";
            default = ugit;
          };

          tsAuthKey = mkOption {
            type = types.str;
            description = "Tailscale one-time auth-key";
            default = "";
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
            args = [
              "--config=${configFile}"
              "--repo-dir=${cfg.repoDir}"
              "--ssh.authorized-keys=${authorizedKeysPath}"
              "--ssh.host-key=${cfg.hostKeyFile}"
            ];
          in "${cfg.package}/bin/ugitd ${builtins.concatStringsSep " " args}";
          wantedBy = ["multi-user.target"];
          after = ["network.target"];
          path = [cfg.package pkgs.git pkgs.bash];
          serviceConfig = {
            User = cfg.user;
            Group = cfg.group;
            Restart = "always";
            RestartSec = "15";
            WorkingDirectory = "/var/lib/ugit";
            Environment = ["TS_AUTHKEY=${cfg.tsAuthKey}"];
          };
        };
      };
    };
  };
}
