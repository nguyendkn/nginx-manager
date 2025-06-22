FROM ubuntu:22.04

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=UTC
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US:en
ENV LC_ALL=en_US.UTF-8

# Create workspace directory
RUN mkdir -p /workspace

# Install basic system dependencies
RUN apt-get update -qq && apt-get install -y --no-install-recommends \
    # Basic tools
    nano \
    vim \
    git \
    curl \
    wget \
    unzip \
    zip \
    jq \
    tree \
    htop \
    # Build tools
    build-essential \
    ca-certificates \
    gnupg \
    lsb-release \
    software-properties-common \
    apt-transport-https \
    # Locales
    locales \
    # Network tools
    net-tools \
    iputils-ping \
    telnet \
    # Process tools
    procps \
    psmisc \
    # SSL/TLS tools
    openssl \
    # Development libraries
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Generate locales
RUN locale-gen en_US.UTF-8

# Set working directory
WORKDIR /workspace

# Create useful directories
RUN mkdir -p /go/src /go/bin /go/pkg \
    && mkdir -p /workspace/logs \
    && mkdir -p /workspace/tmp \
    && mkdir -p /workspace/data

# Install Go 1.24 manually to ensure we have the exact version
ARG GO_VERSION=1.24.4
RUN curl -fsSL https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz | tar -xzC /usr/local \
    && ln -sf /usr/local/go/bin/go /usr/local/bin/go \
    && ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt

# Set Go environment variables
ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$GOROOT/bin:$GOPATH/bin:$PATH
ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org

# Install Node.js 20 manually
ARG NODE_VERSION=20
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs

# Verify installations
RUN go version && node --version && npm --version

# Set up shell environment
RUN echo 'export PS1="\[\033[01;32m\]\u@nginx-manager\[\033[00m\]:\[\033[01;34m\]\w\[\033[00m\]\$ "' >> /root/.bashrc \
    && echo 'export GOROOT=/usr/local/go' >> /root/.bashrc \
    && echo 'export GOPATH=/go' >> /root/.bashrc \
    && echo 'export PATH=$GOROOT/bin:$GOPATH/bin:$PATH' >> /root/.bashrc

# Default command
CMD ["/bin/bash"]
