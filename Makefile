.PHONY: build install clean run help

# Binary name
BINARY_NAME=vgw-manager
INSTALL_PATH=/usr/local/bin

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME)
	@echo "Build complete!"

# Install the binary to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@sudo cp $(BINARY_NAME) $(INSTALL_PATH)/
	@sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@echo "Clean complete!"

# Run the application
run: build
	@./$(BINARY_NAME)

# Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstall complete!"

# Show help
help:
	@echo "Available commands:"
	@echo "  make build      - Build the application"
	@echo "  make install    - Build and install to $(INSTALL_PATH)"
	@echo "  make uninstall  - Remove installed binary"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make run        - Build and run the application"
	@echo "  make help       - Show this help message"
