
# Usage:
# make build all - compile binaries (overwrites existing binaries)
# make clean all - remove all binaries
# make install - install as linux service
# make uninstall - uninstall the linux service

.PHONY: build clean install uninstall

build: 
	cmd/build.sh $(ARGS)

clean:
	cmd/build.sh $(ARGS)

install:
	cmd/install.sh $(ARGS)

uninstall:
	cmd/install.sh --uninstall $(ARGS)
