release:
ifndef VERSION
	$(error VERSION is undefined)
endif
ifndef NEW_DEV_VERSION
	$(error NEW_DEV_VERSION is undefined)
endif
	slu go-code version-bump -v ${VERSION}
	git tag ${VERSION} -a -m ${VERSION}
	slu go-code version-bump -v${NEW_DEV_VERSION}
	git push
	git push --tags