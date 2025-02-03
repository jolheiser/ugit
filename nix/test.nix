{ config, pkgs, ... }:
{
  imports = [ ./module.nix ];

  users.users.jolheiser = {
    isNormalUser = true;
    extraGroups = [ "wheel" ];
    initialPassword = "test";
  };

  services.ugit = {
    enable = true;
    hooks = [
      {
        name = "pre-receive";
        content = ''
          echo "Pre-receive hook executed"
        '';
      }
    ];
  };
}
