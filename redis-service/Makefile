.PHONY: build clean

build:
	@echo "Compiling KCL service configuration..."
	mkdir -p rendered
	kcl run main.k -D output_type=yaml -o rendered/redis-service.yaml

clean:
	rm -rf rendered/*.yaml