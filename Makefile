.PHONY: build build-frontend build-backend clean deploy-frontend generate-cert

ROOT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BUILD_DIR := $(ROOT_DIR)build
FRONTEND_DIR := $(ROOT_DIR)frontend
BACKEND_DIR := $(ROOT_DIR)backend
DEPLOY_DIR := $(ROOT_DIR)deploy

# Deploy host settings — override on the command line, e.g.:
#   make deploy-frontend NGINX_CONF=/etc/nginx/nginx.conf NGINX_WEB_ROOT=/var/www/zscaler-frontend
NGINX_CONF ?= /etc/nginx/nginx.conf
NGINX_WEB_ROOT ?= /var/www/zscaler-frontend
NGINX_SSL_DIR ?= /etc/nginx/ssl

build: build-frontend build-backend

build-frontend:
	@mkdir -p $(BUILD_DIR)/frontend
	cd $(FRONTEND_DIR) && npm run build -- --outDir=$(BUILD_DIR)/frontend --emptyOutDir

build-backend:
	@mkdir -p $(BUILD_DIR)/backend
	cd $(BACKEND_DIR) && go build -o $(BUILD_DIR)/backend/server ./cmd/server

clean:
	rm -rf $(BUILD_DIR)

generate-cert:
	@bash $(DEPLOY_DIR)/generate-cert.sh

deploy-frontend: generate-cert
	@echo "[deploy-frontend] stopping nginx..."
	@-sudo systemctl stop nginx || true
	@echo "[deploy-frontend] installing nginx config to $(NGINX_CONF)..."
	@sudo mkdir -p $(dir $(NGINX_CONF))
	@sudo cp $(DEPLOY_DIR)/nginx.conf $(NGINX_CONF)
	@echo "[deploy-frontend] installing ssl certificates to $(NGINX_SSL_DIR)..."
	@sudo mkdir -p $(NGINX_SSL_DIR)
	@sudo cp $(DEPLOY_DIR)/ssl/zscaler.crt $(NGINX_SSL_DIR)/zscaler.crt
	@sudo cp $(DEPLOY_DIR)/ssl/zscaler.key $(NGINX_SSL_DIR)/zscaler.key
	@sudo chmod 600 $(NGINX_SSL_DIR)/zscaler.key
	@sudo chmod 644 $(NGINX_SSL_DIR)/zscaler.crt
	@echo "[deploy-frontend] copying frontend build to $(NGINX_WEB_ROOT)..."
	@sudo rm -rf $(NGINX_WEB_ROOT)
	@sudo mkdir -p $(NGINX_WEB_ROOT)
	@sudo cp -a $(BUILD_DIR)/frontend/. $(NGINX_WEB_ROOT)/
	@echo "[deploy-frontend] testing nginx config..."
	@sudo nginx -t
	@echo "[deploy-frontend] starting nginx..."
	@sudo systemctl start nginx
	@echo "[deploy-frontend] done. frontend served on https://<host>/"
