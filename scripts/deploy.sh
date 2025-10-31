#!/bin/bash

# BL4 Item Codec - Deployment Script
# This script helps deploy the BL4 API using Docker

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="bl4"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-}"
VERSION="${VERSION:-latest}"
ENVIRONMENT="${ENVIRONMENT:-dev}"

# Functions
print_header() {
    echo -e "${BLUE}=====================================${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}=====================================${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"

    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi
    print_success "Docker is available"

    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_error "Docker Compose is not installed or not in PATH"
        exit 1
    fi
    print_success "Docker Compose is available"

    # Check if we're in the right directory
    if [[ ! -f "go.mod" ]] || [[ ! -f "Dockerfile" ]]; then
        print_error "Please run this script from the project root directory"
        exit 1
    fi
    print_success "Running from project root directory"
}

# Build Docker image
build_image() {
    print_header "Building Docker Image"

    local image_name="${PROJECT_NAME}-api"
    if [[ -n "$DOCKER_REGISTRY" ]]; then
        image_name="${DOCKER_REGISTRY}/${image_name}"
    fi

    print_info "Building image: ${image_name}:${VERSION}"

    docker build -t "${image_name}:${VERSION}" .
    docker tag "${image_name}:${VERSION}" "${image_name}:latest"

    print_success "Docker image built successfully"
}

# Deploy development environment
deploy_dev() {
    print_header "Deploying Development Environment"

    # Stop existing containers
    docker-compose down --remove-orphans 2>/dev/null || true

    # Build and start services
    docker-compose up --build -d

    print_success "Development environment deployed"
    print_info "API is available at: http://localhost:8080"
    print_info "Web interface is available at: http://localhost"
    print_info "Prometheus is available at: http://localhost:9090"
    print_info "Grafana is available at: http://localhost:3000 (admin/admin123)"
}

# Deploy production environment
deploy_prod() {
    print_header "Deploying Production Environment"

    # Check for production configuration
    if [[ ! -f "docker-compose.prod.yml" ]]; then
        print_error "Production configuration file not found"
        exit 1
    fi

    # Check for environment variables
    if [[ -z "$GRAFANA_ADMIN_PASSWORD" ]]; then
        print_warning "GRAFANA_ADMIN_PASSWORD not set, using default (change_me)"
    fi

    # Stop existing containers
    docker-compose -f docker-compose.yml down --remove-orphans 2>/dev/null || true
    docker-compose -f docker-compose.prod.yml down --remove-orphans 2>/dev/null || true

    # Build and start services
    docker-compose -f docker-compose.prod.yml up --build -d

    print_success "Production environment deployed"
    print_info "API is available at: http://localhost:8080"
    print_info "Monitoring available at: http://localhost:9090"
    print_info "Grafana available at: http://localhost:3000"
}

# Health check
health_check() {
    print_header "Performing Health Check"

    local max_attempts=30
    local attempt=1

    while [[ $attempt -le $max_attempts ]]; do
        if curl -f http://localhost:8080/health &>/dev/null; then
            print_success "API is healthy and ready"
            return 0
        fi

        print_info "Waiting for API to be ready... (attempt $attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done

    print_error "API health check failed after $max_attempts attempts"
    return 1
}

# Show logs
show_logs() {
    local service="${1:-}"

    if [[ -n "$service" ]]; then
        docker-compose logs -f "$service"
    else
        docker-compose logs -f
    fi
}

# Stop services
stop_services() {
    print_header "Stopping Services"

    if [[ -f "docker-compose.prod.yml" ]] && docker-compose -f docker-compose.prod.yml ps -q | grep -q .; then
        docker-compose -f docker-compose.prod.yml down
        print_success "Production services stopped"
    fi

    if docker-compose ps -q | grep -q .; then
        docker-compose down
        print_success "Development services stopped"
    fi
}

# Clean up
cleanup() {
    print_header "Cleaning Up"

    # Stop all services
    stop_services

    # Remove images
    local image_name="${PROJECT_NAME}-api"
    if [[ -n "$DOCKER_REGISTRY" ]]; then
        image_name="${DOCKER_REGISTRY}/${image_name}"
    fi

    docker rmi "${image_name}:${VERSION}" 2>/dev/null || true
    docker rmi "${image_name}:latest" 2>/dev/null || true

    # Remove unused volumes
    docker volume prune -f

    print_success "Cleanup completed"
}

# Show help
show_help() {
    cat << EOF
BL4 Item Codec - Deployment Script

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    build           Build Docker image
    dev             Deploy development environment
    prod            Deploy production environment
    health          Perform health check
    logs [service]  Show logs for all services or specific service
    stop            Stop all services
    cleanup         Clean up containers and images
    help            Show this help message

Examples:
    $0 dev                    # Deploy development environment
    $0 prod                   # Deploy production environment
    $0 logs bl4-api          # Show logs for API service
    $0 health                 # Check service health

Environment Variables:
    DOCKER_REGISTRY    Docker registry URL
    VERSION            Image version (default: latest)
    ENVIRONMENT        Environment type (dev/prod)
    GRAFANA_ADMIN_PASSWORD  Grafana admin password (production only)

EOF
}

# Main script logic
main() {
    local command="${1:-help}"

    case "$command" in
        "build")
            check_prerequisites
            build_image
            ;;
        "dev")
            check_prerequisites
            deploy_dev
            sleep 5
            health_check
            ;;
        "prod")
            check_prerequisites
            build_image
            deploy_prod
            sleep 10
            health_check
            ;;
        "health")
            health_check
            ;;
        "logs")
            show_logs "$2"
            ;;
        "stop")
            stop_services
            ;;
        "cleanup")
            cleanup
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Handle script interruption
trap 'print_warning "Script interrupted"; exit 130' INT

# Run main function
main "$@"