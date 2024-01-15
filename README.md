# ugit

<img style="width: 50px;" alt="ugit logo" src="/ugit/tree/main/assets/ugit.svg?raw&pretty"/>

Minimal git server

ugit allows cloning via HTTPS/SSH, but can only be pushed to via SSH.

There are no plans to directly support issues or PR workflows, although webhooks are planned and auxillary software may be created to facilitate these things.
For now, if you wish to collaborate, please send me patches at [john+ugit@jolheiser.com](mailto:john+ugit@jolheiser.com).

Currently all HTML is allowed in markdown, ugit is intended to be run by/for a trusted user.

## Getting your public SSH keys from another forge

Using GitHub as an example (although Gitea/GitLab should have the same URL scheme)

Ba/sh
```sh
curl https://github.com/<username>.keys > path/to/authorized_keys
```

Nushell
```sh
http get https://github.com/<username>.keys | save --force path/to/authorized_keys
```

## License

[MIT](LICENSE)

Lots of inspiration and some starting code used from [wish](https://github.com/charmbracelet/wish) [(MIT)](https://github.com/charmbracelet/wish/blob/3e6f92a166118390484ce4a0904114b375b9e485/LICENSE) and [legit](https://github.com/icyphox/legit) [(MIT)](https://github.com/icyphox/legit/blob/bdfc973207a67a3b217c130520d53373d088763c/license).
