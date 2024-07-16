def main [user = "jolheiser", base_url = "https://git.jolheiser.com", repos = ["ugit", "helix.drv", "tmpl"]] {

  # Clean
  try {
    rm -r .ugit/
    rm -r .ssh/
  }
  
  # SSH
  mkdir .ssh
  http get $"https://github.com/($user).keys" | save --force .ssh/authorized_keys
  
  # Git
  mkdir .ugit
  for $repo in $repos {
    git clone --bare $"($base_url)/($repo).git" $".ugit/($repo).git"
    {"private": false, "description": $repo, "tags": ["git", "dev", "mirror", "archive"]} | save $".ugit/($repo).git/ugit.json"
  }
}
