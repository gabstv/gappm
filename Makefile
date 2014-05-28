build-win:
	gox -osarch="windows/386"
	gox -osarch="windows/amd64"
help:
	@echo make build-win