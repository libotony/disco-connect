all: disco-connect

disco-connect: 
	@echo "building $@..."
	@go build -v -o $(CURDIR)/bin/$@ 
	@echo "done. executable created at 'bin/$@'"

.DEFAULT:
	@$(MAKE) disco-connect