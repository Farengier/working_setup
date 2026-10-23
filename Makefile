build: .urldecode
	
.urldecode:
	@echo "Собираем urldecode"
	VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
	@go build -ldflags "-X main.version=$(VERSION)" -o build/urldecode src/urldecode/main.go

install: build
	@echo "Устанавливаем"
	@mkdir -p ~/bin
	@cp -n ./bin/* ~/bin
	@cp -n ./build/* ~/bin
	@cp -f ./.zshrc ~/