# MetaCRM Development Environment

This directory contains Docker Compose configurations for local development of the MetaCRM application.

## Quick Start

1. **Start all services:**
   ```bash
   cd docker
   docker-compose -f docker-compose.dev.yaml up -d
   ```

2. **View logs:**
   ```bash
   docker-compose -f docker-compose.dev.yaml logs -f
   ```

3. **Stop all services:**
   ```bash
   docker-compose -f docker-compose.dev.yaml down
   ```

## Services

### Core Services

| Service | Port | Description | Credentials |
|---------|------|-------------|-------------|
| **PostgreSQL** | 5432 | Main database | `metacrm` / `metacrm_dev_password` |
| **Redis** | 6379 | Cache and session store | Password: `metacrm_redis_password` |
| **Kafka** | 9092 | Event streaming | No auth required |

### Management UIs

| Service | Port | URL | Credentials |
|---------|------|-----|-------------|
| **pgAdmin** | 8080 | http://localhost:8080 | `admin@metacrm.dev` / `admin` |
| **Redis Commander** | 8081 | http://localhost:8081 | No auth required |
| **Grafana** | 3000 | http://localhost:3000 | `admin` / `admin` |
| **Prometheus** | 9090 | http://localhost:9090 | No auth required |
| **Jaeger** | 16686 | http://localhost:16686 | No auth required |

## Environment Variables

The following environment variables are used by the services:

### PostgreSQL
- `POSTGRES_USER`: Database user (default: `metacrm`)
- `POSTGRES_PASSWORD`: Database password (default: `metacrm_dev_password`)
- `POSTGRES_DB`: Main database name (default: `metacrm`)
- `POSTGRES_DBs`: Comma-separated list of databases to create

### Redis
- Redis is configured with password authentication: `metacrm_redis_password`

### Kafka
- `KAFKA_BROKER_ID`: Kafka broker ID (default: `1`)
- `KAFKA_ZOOKEEPER_CONNECT`: Zookeeper connection string
- `KAFKA_ADVERTISED_LISTENERS`: Kafka advertised listeners

## Development Workflow

### 1. Database Setup
```bash
# Connect to PostgreSQL
docker exec -it metacrm-postgres-dev psql -U metacrm -d metacrm

# Run migrations (from your Go application)
go run cmd/iam/main.go migrate
```

### 2. Redis Operations
```bash
# Connect to Redis CLI
docker exec -it metacrm-redis-dev redis-cli -a metacrm_redis_password

# Test Redis connection
docker exec -it metacrm-redis-dev redis-cli -a metacrm_redis_password ping
```

### 3. Kafka Topics
```bash
# List topics
docker exec -it metacrm-kafka-dev kafka-topics --bootstrap-server localhost:9092 --list

# Create a topic
docker exec -it metacrm-kafka-dev kafka-topics --bootstrap-server localhost:9092 --create --topic iam.events --partitions 3 --replication-factor 1
```

### 4. Monitoring
- **Grafana**: http://localhost:3000 - Metrics dashboards
- **Prometheus**: http://localhost:9090 - Metrics collection
- **Jaeger**: http://localhost:16686 - Distributed tracing

## Service Dependencies

The services have the following startup dependencies:
- `kafka` depends on `zookeeper`
- `redis-commander` depends on `redis`
- `pgadmin` depends on `postgres`

## Health Checks

All services include health checks to ensure they're ready before dependent services start:
- PostgreSQL: `pg_isready` check
- Redis: `redis-cli ping` check
- Kafka: `kafka-broker-api-versions` check

## Data Persistence

The following data is persisted in Docker volumes:
- `postgres_data`: PostgreSQL database files
- `redis_data`: Redis AOF (Append Only File) data
- `pgadmin_data`: pgAdmin configuration and data
- `grafana_data`: Grafana dashboards and configuration

## Troubleshooting

### Common Issues

1. **Port conflicts**: If you get port binding errors, check if other services are using the same ports
2. **Permission issues**: On Linux, you might need to run with `sudo` or add your user to the docker group
3. **Memory issues**: Ensure Docker has enough memory allocated (recommended: 4GB+)

### Useful Commands

```bash
# Check service status
docker-compose -f docker-compose.dev.yaml ps

# View service logs
docker-compose -f docker-compose.dev.yaml logs [service-name]

# Restart a specific service
docker-compose -f docker-compose.dev.yaml restart [service-name]

# Remove all containers and volumes (WARNING: This will delete all data)
docker-compose -f docker-compose.dev.yaml down -v

# Rebuild and restart services
docker-compose -f docker-compose.dev.yaml up -d --build
```

## Integration with Go Services

To connect your Go services to these containers, use the following connection strings:

### PostgreSQL
```
postgres://metacrm:metacrm_dev_password@localhost:5432/metacrm?sslmode=disable
```

### Redis
```
redis://:metacrm_redis_password@localhost:6379/0
```

### Kafka
```
localhost:9092
```

## Security Note

⚠️ **This configuration is for development only!** 
- Default passwords are used
- No SSL/TLS encryption
- Services are exposed on localhost
- Do not use these credentials in production
