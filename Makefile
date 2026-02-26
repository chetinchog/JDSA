BINARY_NAME=jdsa
RELEASE_DIR=release
BUILD_BIN_DIR=build/bin

.PHONY: all build run clean

all: build

build:
	@echo "[1/4] Creando carpeta de release y build..."
	@mkdir -p $(RELEASE_DIR)
	@mkdir -p $(BUILD_BIN_DIR)
	@echo "[2/4] Cerrando instancia de $(BINARY_NAME) si esta abierta..."
	-@pkill -if $(BINARY_NAME) 2>/dev/null || true
	@sleep 1
	@echo "[3/4] Compilando aplicacion con Wails..."
	@which wails > /dev/null || (echo "Error: 'wails' command not found. Please install it or add it to your PATH." && exit 1)
	@wails build -o $(BINARY_NAME)
	@echo "[4/4] Moviendo ejecutable a la carpeta de release..."
	@rm -f $(RELEASE_DIR)/$(BINARY_NAME)
	@if [ -f $(BUILD_BIN_DIR)/$(BINARY_NAME) ]; then \
		mv $(BUILD_BIN_DIR)/$(BINARY_NAME) $(RELEASE_DIR)/$(BINARY_NAME); \
	elif [ -d $(BUILD_BIN_DIR)/$(BINARY_NAME).app ]; then \
		cp $(BUILD_BIN_DIR)/$(BINARY_NAME).app/Contents/MacOS/$(BINARY_NAME) $(RELEASE_DIR)/$(BINARY_NAME); \
	else \
		echo "Error: Binary not found in $(BUILD_BIN_DIR)"; \
		exit 1; \
	fi
	@echo "\n[EXITO] Build completado: ./$(RELEASE_DIR)/$(BINARY_NAME)"

run: build
	@echo "[5/5] Ejecutando $(BINARY_NAME)..."
	@./$(RELEASE_DIR)/$(BINARY_NAME)
	@echo "\n[EXITO] Listo!"

clean:
	@echo "Limpiando..."
	@rm -rf $(RELEASE_DIR)
	@rm -rf build/bin
