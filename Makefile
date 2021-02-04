
# Usage:
# make build all - compile binaries (overwrites existing binaries)
# make clean all - remove all binaries
# make install - install as linux service
# make uninstall - uninstall the linux service

.PHONY: build clean install uninstall

build: go
	scripts/build.sh $(ARGS)

clean:
	scripts/build.sh $(ARGS)

install:
	scripts/install.sh $(ARGS)

uninstall:
	scripts/install.sh --uninstall $(ARGS)
