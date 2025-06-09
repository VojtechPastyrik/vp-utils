# vp-utils
My CLI utils

# Install
MacOS:
```
brew install VojtechPastyrik/tap/vp-utils
```

Ubuntu/Debian:

```
curl -1sLf \
  'https://dl.cloudsmith.io/public/vojtechpastyrik/vp-utils/setup.deb.sh' \
  | sudo -E bash
sudo apt update
sudo apt install vp-utils
```

Fedora:

```
curl -1sLf \
  'https://dl.cloudsmith.io/public/vojtechpastyrik/vp-utils/setup.rpm.sh' \
  | sudo -E bash
sudo dnf install vp-utils
```

# Docker

You can also run vp-utils in a Docker container. The image is available on Docker Hub:
see [vojtechpastyrik/vp-utils](https://hub.docker.com/r/vojtechpastyrik/vp-utils).
Image is built automatically from the latest release, so you can use it without worrying about updates. You can find two
versions of the image: amd64 and arm64.
To run the container, use the following command:

```bash
docker run --rm vojtechpastyrik/vp-utils:latest vp-utils
```

# Release

Update version in `version/version.go` using [slu](https://github.com/sikalabs/slu), create new tag and push it.

```shell
make release VERSION=v0.7.1 NEW_DEV_VERSION=v0.8.0-dev
```
