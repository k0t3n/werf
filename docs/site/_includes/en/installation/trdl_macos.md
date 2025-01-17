Make sure you have Git 2.18.0 or newer and [Docker](https://docs.docker.com/get-docker) installed.

Setup [trdl](https://github.com/werf/trdl) which will manage `werf` installation and updates:
```shell
# Add ~/bin to the PATH.
echo 'export PATH=$HOME/bin:$PATH' >> ~/.zprofile
export PATH="$HOME/bin:$PATH"

# Install trdl.
curl -L "https://tuf.trdl.dev/targets/releases/0.1.3/darwin-{{ include.arch }}/bin/trdl" -o /tmp/trdl
mkdir -p ~/bin
install /tmp/trdl ~/bin/trdl
```

Add `werf` repo:
```shell
trdl add werf https://tuf.werf.io 1 b7ff6bcbe598e072a86d595a3621924c8612c7e6dc6a82e919abe89707d7e3f468e616b5635630680dd1e98fc362ae5051728406700e6274c5ed1ad92bea52a2
```

For local usage we recommend automatically activating `werf` for new shell sessions:
```shell
echo 'source $(trdl use werf {{ include.version }} {{ include.channel }})' >> ~/.zshrc
```

But in CI you should prefer activating `werf` explicitly in the beginning of each job/pipeline:
```shell
source $(trdl use werf {{ include.version }} {{ include.channel }})
```

Make sure that `werf` is available now (open new shell if you chose automatic activation):
```shell
werf version
```
