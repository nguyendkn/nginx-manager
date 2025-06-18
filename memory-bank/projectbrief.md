# Project Brief: Nginx Manager

## Project Overview

**Project Name**: nginx-manager
**Repository**: github.com/nguyendkn/nginx-manager
**Backend**: Go 1.24.4
**Frontend**: React 19.1.0 with TypeScript and React Router v7
**Architecture**: Full-stack web application with REST API and modern web UI

## Purpose

Nginx Manager is a comprehensive web-based system for managing Nginx server configurations through a REST API. The project aims to simplify the complex task of Nginx configuration management by providing automated tools for configuration creation, validation, deployment, and monitoring.

## Core Goals

1. **Configuration Management**: Provide CRUD operations for Nginx configuration files
2. **Validation**: Ensure configuration validity before deployment
3. **Service Control**: Manage Nginx service lifecycle (start, stop, reload, restart)
4. **Health Monitoring**: Monitor Nginx server health and performance metrics
5. **Multi-Interface Access**: Support both REST API and CLI interfaces
6. **Automation**: Enable scheduled tasks through cronjob functionality

## Target Users

- **DevOps Engineers**: Need efficient tools for managing multiple Nginx instances
- **System Administrators**: Require safe, validated configuration management
- **Development Teams**: Want automated deployment of Nginx configurations
- **Operations Teams**: Need monitoring and health check capabilities

## Key Requirements

### Functional Requirements
- REST API for all Nginx management operations
- Configuration file validation before deployment
- Real-time health monitoring and status reporting
- CLI tool for command-line operations
- Scheduled task execution for automated maintenance
- Comprehensive logging and audit trails

### Non-Functional Requirements
- **Reliability**: Safe configuration changes with rollback capabilities
- **Performance**: Fast API responses and efficient resource usage
- **Security**: Secure access controls and input validation
- **Maintainability**: Clean architecture with modular components
- **Observability**: Comprehensive logging and monitoring

## Architecture Components

1. **HTTP Server** (`cmd/server`): REST API service for backend operations
2. **Web UI** (`webui/`): Modern React-based frontend application
3. **CLI Tool** (`cmd/cli`): Command-line interface for direct operations
4. **Cronjob Service** (`cmd/cronjob`): Automated task execution
5. **Core Logic** (`internal/`): Business logic and controllers
6. **Shared Packages** (`pkg/`): Reusable utilities and libraries

## Success Criteria

- Reduce Nginx configuration deployment time by 80%
- Eliminate configuration syntax errors through validation
- Provide 99.9% uptime for the management service
- Enable zero-downtime Nginx configuration updates
- Support management of multiple Nginx instances from single interface

## Project Constraints

- Must be compatible with standard Nginx installations
- Should not require modifications to existing Nginx setups
- Must maintain backward compatibility with existing configurations
- Should be deployable in containerized environments
- Must support both local and remote Nginx management
