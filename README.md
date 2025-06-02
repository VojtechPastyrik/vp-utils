# vp-utils
My CLI utils

# Install
MacOS:
```
brew install VojtechPastyrik/tap/vp-utils
```


## Release

Update version in `version/version.go` using [slu](https://github.com/sikalabs/slu), create new tag and push it.

```
VERSION=v0.16.0
NEW_DEV_VERSION=v0.17.0-dev
slu go-code version-bump -v $VERSION
git tag $VERSION -a -m $VERSION
slu go-code version-bump -v $NEW_DEV_VERSION
git push
git push --tags
```
