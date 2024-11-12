.PHONY: build run test clean

include .env
export $(shell sed 's/=.*//' .env)

APP_NAME=eth_bal
BUILD_DIR=./bin
MAIN_FILE=./cmd/app/main.go

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)

run: build
	$(BUILD_DIR)/$(APP_NAME)

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)/*
