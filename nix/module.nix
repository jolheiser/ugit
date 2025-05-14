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
  instanceOptions =
    { name, config, ... }:
    let
      inherit (lib) mkEnableOption mkOption types;
      baseDir = "/var/lib/ugit-${name}";
    in
    {
      options = {
        enable = mkEnableOption "Enable ugit";

        package = mkOption {
          type = types.package;
          description = "ugit package to use";
          default = pkg;
        };

        homeDir = mkOption {
          type = types.str;
          description = "ugit home directory";
          default = baseDir;
        };

        repoDir = mkOption {
          type = types.str;
          description = "where ugit stores repositories";
          default = "${baseDir}/repos";
        };

        authorizedKeys = mkOption {
          type = types.listOf types.str;
          description = "list of keys to use for authorized_keys";
          default = [ ];
        };

        authorizedKeysFile = mkOption {
          type = types.str;
          description = "path to authorized_keys file ugit uses for auth";
          default = "${baseDir}/authorized_keys";
        };

        hostKeyFile = mkOption {
          type = types.str;
          description = "path to host key file (will be created if it doesn't exist)";
          default = "${baseDir}/ugit_ed25519";
        };

        config = mkOption {
          type = types.attrs;
          default = { };
          description = "config.yaml contents";
        };

        user = mkOption {
          type = types.str;
          default = "ugit-${name}";
          description = "User account under which ugit runs";
        };

        group = mkOption {
          type = types.str;
          default = "ugit-${name}";
          description = "Group account under which ugit runs";
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
in
{
  options = {
    services.ugit = lib.mkOption {
      type = lib.types.attrsOf (lib.types.submodule instanceOptions);
      default = { };
      description = "Attribute set of ugit instances";
    };
  };
  config = lib.mkIf (cfg != { }) {
    users.users = lib.mapAttrs' (
      name: instanceCfg:
      lib.nameValuePair instanceCfg.user {
        home = instanceCfg.homeDir;
        createHome = true;
        group = instanceCfg.group;
        isSystemUser = true;
        isNormalUser = false;
        description = "user for ugit ${name} service";
      }
    ) (lib.filterAttrs (name: instanceCfg: instanceCfg.enable) cfg);

    users.groups = lib.mapAttrs' (name: instanceCfg: lib.nameValuePair instanceCfg.group { }) (
      lib.filterAttrs (name: instanceCfg: instanceCfg.enable) cfg
    );

    systemd.services = lib.foldl' (
      acc: name:
      let
        instanceCfg = cfg.${name};
      in
      lib.recursiveUpdate acc (
        lib.optionalAttrs instanceCfg.enable {
          "ugit-${name}" = {
            enable = true;
            description = "ugit instance ${name}";
            wantedBy = [ "multi-user.target" ];
            after = [ "network.target" ];
            path = [
              instanceCfg.package
              pkgs.git
              pkgs.bash
            ];
            serviceConfig = {
              User = instanceCfg.user;
              Group = instanceCfg.group;
              Restart = "always";
              RestartSec = "15";
              WorkingDirectory = instanceCfg.homeDir;
              ReadWritePaths = [ instanceCfg.homeDir ];
              CapabilityBoundingSet = "";
              NoNewPrivileges = true;
              ProtectSystem = "strict";
              ProtectHome = true;
              PrivateTmp = true;
              PrivateDevices = true;
              PrivateUsers = true;
              ProtectHostname = true;
              ProtectClock = true;
              ProtectKernelTunables = true;
              ProtectKernelModules = true;
              ProtectKernelLogs = true;
              ProtectControlGroups = true;
              RestrictAddressFamilies = [
                "AF_UNIX"
                "AF_INET"
                "AF_INET6"
              ];
              RestrictNamespaces = true;
              LockPersonality = true;
              MemoryDenyWriteExecute = true;
              RestrictRealtime = true;
              RestrictSUIDSGID = true;
              RemoveIPC = true;
              PrivateMounts = true;
              SystemCallArchitectures = "native";
              ExecStart =
                let
                  configFile = pkgs.writeText "ugit-${name}.yaml" (
                    builtins.readFile (yamlFormat.generate "ugit-${name}-yaml" instanceCfg.config)
                  );
                  authorizedKeysFile = pkgs.writeText "ugit_${name}_keys" (
                    builtins.concatStringsSep "\n" instanceCfg.authorizedKeys
                  );

                  authorizedKeysPath =
                    if (builtins.length instanceCfg.authorizedKeys) > 0 then
                      authorizedKeysFile
                    else
                      instanceCfg.authorizedKeysFile;
                  args = [
                    "--config=${configFile}"
                    "--repo-dir=${instanceCfg.repoDir}"
                    "--ssh.authorized-keys=${authorizedKeysPath}"
                    "--ssh.host-key=${instanceCfg.hostKeyFile}"
                  ];
                in
                "${instanceCfg.package}/bin/ugitd ${builtins.concatStringsSep " " args}";
            };
          };

          "ugit-${name}-hooks" = {
            description = "Setup hooks for ugit instance ${name}";
            wantedBy = [ "multi-user.target" ];
            after = [ "ugit-${name}.service" ];
            requires = [ "ugit-${name}.service" ];
            serviceConfig = {
              Type = "oneshot";
              RemainAfterExit = true;
              User = instanceCfg.user;
              Group = instanceCfg.group;
              ExecStart =
                let
                  hookDir = "${instanceCfg.repoDir}/hooks/pre-receive.d";
                  mkHookScript =
                    hook:
                    let
                      script = pkgs.writeShellScript "ugit-${name}-${hook.name}" hook.content;
                    in
                    ''
                      mkdir -p ${hookDir}
                      ln -sf ${script} ${hookDir}/${hook.name}
                    '';
                in
                pkgs.writeShellScript "ugit-${name}-hooks-setup" ''
                  ${builtins.concatStringsSep "\n" (map mkHookScript instanceCfg.hooks)}
                '';
            };
          };
        }
      )
    ) { } (builtins.attrNames cfg);
  };
}
