{ pkgs, ... }:
let
  privKey = ''
    -----BEGIN OPENSSH PRIVATE KEY-----
    b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
    QyNTUxOQAAACBIpmLtcHhECei1ls6s0kKUehjpRCP9yel/c5YCIb5DpQAAAIgAYtkzAGLZ
    MwAAAAtzc2gtZWQyNTUxOQAAACBIpmLtcHhECei1ls6s0kKUehjpRCP9yel/c5YCIb5DpQ
    AAAEDFY3M69VfnFbyE67r3l4lDcf5eht5qgNemE9xtMhRkBkimYu1weEQJ6LWWzqzSQpR6
    GOlEI/3J6X9zlgIhvkOlAAAAAAECAwQF
    -----END OPENSSH PRIVATE KEY-----
  '';
  pubKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEimYu1weEQJ6LWWzqzSQpR6GOlEI/3J6X9zlgIhvkOl";
  sshConfig = ''
    Host ugit
        HostName localhost
        Port 8448
        User ugit
        IdentityFile ~/.ssh/vm
        IdentitiesOnly yes
  '';
in
{
  imports = [ ./module.nix ];
  environment.systemPackages = with pkgs; [ git ];
  services.getty.autologinUser = "root";
  services.openssh.enable = true;
  services.ugit.vm = {
    enable = true;
    authorizedKeys = [ pubKey ];
    hooks = [
      {
        name = "pre-receive";
        content = ''
          echo "Pre-receive hook executed"
        '';
      }
    ];
  };
  systemd.services."setup-vm" = {
    wantedBy = [ "multi-user.target" ];
    after = [ "ugit-vm.service" ];
    path = with pkgs; [
      git
    ];
    serviceConfig = {
      Type = "oneshot";
      RemainAfterExit = true;
      User = "root";
      Group = "root";
      ExecStart =
        let
          privSSH = pkgs.writeText "vm-privkey" privKey;
          sshConfigFile = pkgs.writeText "vm-sshconfig" sshConfig;
        in
        pkgs.writeShellScript "setup-vm-script" ''
          # Hack to let ugit start up and generate its SSH keypair
          sleep 3

          # Set up git
          git config --global user.name "NixUser"
          git config --global user.email "nixuser@example.com"
          git config --global init.defaultBranch main
          git config --global push.autoSetupRemote true

          # Set up SSH files
          mkdir ~/.ssh
          ln -sf ${sshConfigFile} ~/.ssh/config
          cp ${privSSH} ~/.ssh/vm
          chmod 600 ~/.ssh/vm
          echo "[localhost]:8448 $(cat /var/lib/ugit-vm/ugit_ed25519.pub)" > ~/.ssh/known_hosts

          # Stage some git activity
          mkdir ~/repo
          cd ~/repo
          git init
          git remote add origin ugit:repo.git
          touch README.md
          git add README.md
          git commit -m "Test"
        '';
    };
  };

}
