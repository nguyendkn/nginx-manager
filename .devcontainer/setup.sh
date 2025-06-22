#!/bin/bash

# Setup script for Nginx Manager development container
set -e

echo "🚀 Setting up Nginx Manager development environment..."

# Update package lists
echo "📦 Updating package lists..."
apt-get update -qq

# Install additional system dependencies
echo "🔧 Installing system dependencies..."
apt-get install -y --no-install-recommends \
    build-essential \
    ca-certificates \
    gnupg \
    lsb-release \
    software-properties-common \
    apt-transport-https \
    unzip \
    jq \
    tree \
    htop \
    vim \
    postgresql-client \
    redis-tools

# Verify Go installation
echo "🐹 Verifying Go installation..."
go version

# Install Go tools
echo "🔧 Installing Go development tools..."
go install -v golang.org/x/tools/gopls@latest
go install -v github.com/go-delve/delve/cmd/dlv@latest
go install -v honnef.co/go/tools/cmd/staticcheck@latest
go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install -v github.com/swaggo/swag/cmd/swag@latest
go install -v github.com/air-verse/air@latest

# Verify Node.js installation
echo "📦 Verifying Node.js installation..."
node --version
npm --version

# Install Node.js tools globally
echo "🔧 Installing Node.js development tools..."
npm install -g \
    typescript \
    @types/node \
    eslint \
    prettier \
    @vitejs/plugin-react \
    concurrently \
    nodemon

# Set up Go workspace
echo "🏗️ Setting up Go workspace..."
mkdir -p /go/src /go/bin /go/pkg
export GOPATH=/go
export PATH=$GOPATH/bin:$PATH

# Install project dependencies
echo "📥 Installing project dependencies..."

# Go dependencies
if [ -f "/workspace/go.mod" ]; then
    cd /workspace
    echo "Installing Go dependencies..."
    go mod download
    go mod tidy
fi

# Node.js dependencies
if [ -f "/workspace/webui/package.json" ]; then
    cd /workspace/webui
    echo "Installing Node.js dependencies..."
    npm install
fi

# Set up Git configuration (if not already configured)
echo "🔧 Setting up Git configuration..."
if [ ! -f ~/.gitconfig ]; then
    git config --global init.defaultBranch main
    git config --global core.autocrlf input
    git config --global pull.rebase false
fi

# Create useful aliases
echo "⚡ Setting up shell aliases..."
cat >> ~/.bashrc << 'EOF'

# Nginx Manager Development Aliases
alias ll='ls -alF'
alias la='ls -A'
alias l='ls -CF'
alias ..='cd ..'
alias ...='cd ../..'

# Go aliases
alias gob='go build'
alias gor='go run'
alias got='go test'
alias gom='go mod'
alias gov='go version'
alias goi='go install'
alias goc='go clean'
alias gof='go fmt'
alias govet='go vet'

# Git aliases
alias gs='git status'
alias ga='git add'
alias gc='git commit'
alias gp='git push'
alias gl='git log --oneline'

# Project specific
alias backend='cd /workspace && go run cmd/server/main.go'
alias frontend='cd /workspace/webui && npm run dev'
alias dev='cd /workspace && concurrently "go run cmd/server/main.go" "cd webui && npm run dev"'

# Docker aliases
alias dc='docker-compose'
alias dps='docker ps'
alias di='docker images'

EOF

# Set up environment variables
echo "🌍 Setting up environment variables..."
cat >> ~/.bashrc << 'EOF'

# Go environment
export GOPATH=/go
export PATH=$GOPATH/bin:$PATH
export GO111MODULE=on

# Node environment
export NODE_ENV=development

# Project environment
export NGINX_MANAGER_ENV=development

EOF

# Create development directories
echo "📁 Creating development directories..."
mkdir -p /workspace/logs
mkdir -p /workspace/tmp
mkdir -p /workspace/data

# Set proper permissions
echo "🔐 Setting permissions..."
chown -R root:root /workspace
chmod -R 755 /workspace

# Clean up
echo "🧹 Cleaning up..."
apt-get autoremove -y
apt-get autoclean
rm -rf /var/lib/apt/lists/*

echo "✅ Development environment setup complete!"
echo ""
echo "🔍 Installed versions:"
go version
node --version
npm --version
echo ""
echo "🎯 Quick start commands:"
echo "  - Start backend:  go run cmd/server/main.go"
echo "  - Start frontend: cd webui && npm run dev"
echo "  - Start both:     dev"
echo "  - Run tests:      go test ./..."
echo "  - Build project:  go build -o nginx-manager cmd/server/main.go"
echo "  - Hot reload:     air (if available)"
echo ""
echo "🔗 Useful ports:"
echo "  - Backend API:    http://localhost:8080"
echo "  - Frontend:       http://localhost:5173"
echo "  - PostgreSQL:     localhost:5432"
echo "  - Redis:          localhost:6379"
echo ""
echo "🛠️ Go tools available:"
echo "  - gopls (language server)"
echo "  - dlv (debugger)"
echo "  - staticcheck (linter)"
echo "  - golangci-lint (meta-linter)"
echo "  - swag (swagger generator)"
echo "  - air (hot reload)"
echo ""
echo "Happy coding with Go 1.24! 🚀"
