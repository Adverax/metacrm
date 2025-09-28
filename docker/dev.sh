#!/bin/bash

# MetaCRM Development Environment Helper Script

set -e

COMPOSE_FILE="docker-compose.dev.yaml"
PROJECT_NAME="metacrm-dev"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if docker-compose is available
check_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        log_error "docker-compose is not installed or not in PATH"
        exit 1
    fi
}

# Start all services
start_services() {
    log_info "Starting MetaCRM development environment..."
    docker-compose -f $COMPOSE_FILE up -d
    log_success "All services started successfully!"
    
    echo ""
    log_info "Service URLs:"
    echo "  PostgreSQL: localhost:5432"
    echo "  Redis: localhost:6379"
    echo "  Kafka: localhost:9092"
    echo "  pgAdmin: http://localhost:8080"
    echo "  Redis Commander: http://localhost:8081"
    echo "  Grafana: http://localhost:3000"
    echo "  Prometheus: http://localhost:9090"
    echo "  Jaeger: http://localhost:16686"
}

# Stop all services
stop_services() {
    log_info "Stopping MetaCRM development environment..."
    docker-compose -f $COMPOSE_FILE down
    log_success "All services stopped successfully!"
}

# Restart all services
restart_services() {
    log_info "Restarting MetaCRM development environment..."
    docker-compose -f $COMPOSE_FILE restart
    log_success "All services restarted successfully!"
}

# Show service status
show_status() {
    log_info "Service status:"
    docker-compose -f $COMPOSE_FILE ps
}

# Show logs
show_logs() {
    if [ -n "$1" ]; then
        log_info "Showing logs for service: $1"
        docker-compose -f $COMPOSE_FILE logs -f "$1"
    else
        log_info "Showing logs for all services:"
        docker-compose -f $COMPOSE_FILE logs -f
    fi
}

# Clean up (remove containers and volumes)
cleanup() {
    log_warning "This will remove all containers and volumes. All data will be lost!"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Cleaning up MetaCRM development environment..."
        docker-compose -f $COMPOSE_FILE down -v
        docker system prune -f
        log_success "Cleanup completed!"
    else
        log_info "Cleanup cancelled."
    fi
}

# Test connections
test_connections() {
    log_info "Testing service connections..."
    
    # Test PostgreSQL
    if docker exec metacrm-postgres-dev pg_isready -U metacrm -d metacrm > /dev/null 2>&1; then
        log_success "PostgreSQL: ✓ Connected"
    else
        log_error "PostgreSQL: ✗ Connection failed"
    fi
    
    # Test Redis
    if docker exec metacrm-redis-dev redis-cli -a metacrm_redis_password ping > /dev/null 2>&1; then
        log_success "Redis: ✓ Connected"
    else
        log_error "Redis: ✗ Connection failed"
    fi
    
    # Test Kafka
    if docker exec metacrm-kafka-dev kafka-broker-api-versions --bootstrap-server localhost:9092 > /dev/null 2>&1; then
        log_success "Kafka: ✓ Connected"
    else
        log_error "Kafka: ✗ Connection failed"
    fi
}

# Show help
show_help() {
    echo "MetaCRM Development Environment Helper"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  start       Start all services"
    echo "  stop        Stop all services"
    echo "  restart     Restart all services"
    echo "  status      Show service status"
    echo "  logs [svc]  Show logs (optionally for specific service)"
    echo "  test        Test service connections"
    echo "  cleanup     Remove all containers and volumes (DESTRUCTIVE)"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 start"
    echo "  $0 logs postgres"
    echo "  $0 test"
}

# Main script logic
main() {
    check_docker_compose
    
    case "${1:-help}" in
        start)
            start_services
            ;;
        stop)
            stop_services
            ;;
        restart)
            restart_services
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs "$2"
            ;;
        test)
            test_connections
            ;;
        cleanup)
            cleanup
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
