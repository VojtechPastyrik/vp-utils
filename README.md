# vpd
My CLI utils

# Install
MacOS:
```
brew install VojtechPastyrik/tap/vpd
```

Ubuntu/Debian:

```
curl -1sLf \
  'https://dl.cloudsmith.io/public/vojtechpastyrik/vpd/setup.deb.sh' \
  | sudo -E bash
sudo apt update
sudo apt install vpd
```

Fedora:

```
curl -1sLf \
  'https://dl.cloudsmith.io/public/vojtechpastyrik/vpd/setup.rpm.sh' \
  | sudo -E bash
sudo dnf install vpd
```

# Docker

You can also run vpd in a Docker container. The image is available on Docker Hub:
see [vojtechpastyrik/vpd](https://hub.docker.com/r/vojtechpastyrik/vpd).
Image is built automatically from the latest release, so you can use it without worrying about updates. You can find two
versions of the image: amd64 and arm64.
To run the container, use the following command:

```bash
docker run --rm vojtechpastyrik/vpd:latest vpd
```

# Release

Update version in `version/version.go` using `vpd release`, create new tag and push it.