.PHONY: build clean

build:
	@echo "Compiling KCL platform configuration..."
	mkdir -p rendered
	kcl run main.k -o rendered/platform.yaml

clean:
	rm -rf rendered/*.yaml