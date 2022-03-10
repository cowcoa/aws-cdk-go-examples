.DEFAULT_GOAL := build

BUILD_ENV_FLAGS := GOARCH=arm64 GOOS=linux
TARGET_DIR := ${CURDIR}/bin

build:
	@for target in $(shell ls -Ibin -I*file -I*.sh); do \
		pushd $$target &> /dev/null; \
		$(BUILD_ENV_FLAGS) go build -o $(TARGET_DIR)/$$target; \
		popd &> /dev/null; \
	done

clean:
	@rm -rf $(TARGET_DIR)
