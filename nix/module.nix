{
  pkgs,
  lib,
  config,
  ...
}:
let
  cfg = config.services.ugit;
  pkg = pkgs.callPackage ./pkg.nix { inherit pkgs; };
  yamlFormat = pkgs.formats.yaml { };
  configFile = pkgs.writeText "ugit.yaml" (
    builtins.readFile (yamlFormat.generate "ugit-yaml" cfg.config)
  );
  authorizedKeysFile = pkgs.writeText "ugit_keys" (builtins.concatStringsSep "\n" cfg.authorizedKeys);
in
{
  options =
    let
      inherit (lib) mkEnableOption mkOption types;
    in
    {
      services.ugit = {
        enable = mkEnableOption "Enable ugit";

        package = mkOption {
          type = types.package;
          description = "ugit package to use";
          default = pkg;
        };

        repoDir = mkOption {
          type = types.str;
          description = "where ugit stores repositories";
          default = "/var/lib/ugit/repos";
        };

        authorizedKeys = mkOption {
          type = types.listOf types.str;
          description = "list of keys to use for authorized_keys";
          default = [ ];
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
          default = { };
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

        hooks = mkOption {
          type = types.listOf (
            types.submodule {
              options = {
                name = mkOption {
                  type = types.str;
                  description = "Hook name";
                };
                content = mkOption {
                  type = types.str;
                  description = "Hook contents";
                };
              };
            }
          );
          description = "A list of pre-receive hooks to run";
          default = [ ];
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
    users.groups."${cfg.group}" = { };
    networking.firewall = lib.mkIf cfg.openFirewall {
      allowedTCPPorts = [
        8448
        8449
      ];
    };

    systemd.services = {
      ugit = {
        enable = true;
        script =
          let
            authorizedKeysPath =
              if (builtins.length cfg.authorizedKeys) > 0 then authorizedKeysFile else cfg.authorizedKeysFile;
            args = [
              "--config=${configFile}"
              "--repo-dir=${cfg.repoDir}"
              "--ssh.authorized-keys=${authorizedKeysPath}"
              "--ssh.host-key=${cfg.hostKeyFile}"
            ];
          in
          "${cfg.package}/bin/ugitd ${builtins.concatStringsSep " " args}";
        wantedBy = [ "multi-user.target" ];
        after = [ "network.target" ];
        path = [
          cfg.package
          pkgs.git
          pkgs.bash
        ];
        serviceConfig = {
          User = cfg.user;
          Group = cfg.group;
          Restart = "always";
          RestartSec = "15";
          WorkingDirectory = "/var/lib/ugit";
        };
      };
      ugit-hooks = {
        wantedBy = [ "multi-user.target" ];
        after = [ "ugit.service" ];
        requires = [ "ugit.service" ];
        serviceConfig = {
          Type = "oneshot";
          ExecStart =
            let
              script = pkgs.writeShellScript "ugit-hooks-link" (
                builtins.concatStringsSep "\n" (
                  map (
                    hook:
                    let
                      script = pkgs.writeShellScript hook.name hook.content;
                      path = "${cfg.repoDir}/hooks/pre-receive.d/${hook.name}";
                    in
                    "ln -s ${script} ${path}"
                  ) cfg.hooks
                )
              );
            in
            "${script}";
        };
      };
    };

    systemd.tmpfiles.settings.ugit = builtins.listToAttrs (
      map (
        hook:
        let
          script = pkgs.writeShellScript hook.name hook.content;
          path = "${cfg.repoDir}/hooks/pre-receive.d/${hook.name}";
        in
        {
          name = path;
          value = {
            "L" = {
              argument = "${script}";
            };
          };
        }
      ) cfg.hooks
    );
  };
}
