# Target name (executable file)
TARGET = server

# Paths and variables
SRC_DIR = ../cmd/server
BUILD_DIR = ./
GO = go
PROTOC = protoc

# Generate code for protobuf
.PHONY: protobuf
protobuf:
	@echo "Generating code for protobuf..."
	$(PROTOC) -I=../protocol --go_out=../internal ../protocol/p2p.proto


# Cleaning generated files
.PHONY: clean
clean:
	@echo "Cleaning the executable file..."
	@rm -f $(TARGET)

# Building the executable file
.PHONY: build
build: clean
	@echo "Building the executable file..."
	$(GO) mod download
	$(GO) build $(BUILD_FLAGS) -o $(TARGET) $(SRC_DIR)

# Running the built executable file
.PHONY: run
run: build
	@echo "Running the executable file..."
	./$(TARGET) -port $(PORT)

# Help (display available commands)
.PHONY: help
help:
	@echo "Available commands:"
	@echo "    make protobuf"
	@echo "        Generate code for protobuf"
	@echo "    make clean"
	@echo "        Clean generated files"
	@echo "    make build"
	@echo "        Build the executable file"
	@echo "    make run"
	@echo "        Run the built executable file"
	@echo "    make help"
	@echo "        Display this message"

# Default action is to display help
.DEFAULT_GOAL := help
