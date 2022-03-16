.DEFAULT_GOAL := build

GO_ARCH := arm64
BUILD_ENV_FLAGS := GOARCH=$(GO_ARCH) GOOS=linux
TARGET_DIR := ${CURDIR}/bin

build:
	@for target in $(shell ls -Ibin -IDockerfile* -IMakefile -I*.sh); do \
		pushd $$target &> /dev/null; \
		$(BUILD_ENV_FLAGS) go build -o $(TARGET_DIR)/$$target; \
		popd &> /dev/null; \
	done
