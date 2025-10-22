# RabbitMQ Local Setup Guide

## Quick Start

### Start RabbitMQ Container

```bash
# Start RabbitMQ in the background
docker-compose up -d

# Check if it's running
docker-compose ps

# View logs
docker-compose logs -f rabbitmq
```

### Access RabbitMQ Management UI

Open in your browser:
- **URL**: http://localhost:15672
- **Username**: guest
- **Password**: guest

### Run Load Tests

```bash
# Light load test (30 seconds)
go run ./main.go rabbitmq load-test --profile light

# Medium load test (2 minutes)
go run ./main.go rabbitmq load-test --profile medium

# Heavy load test (3 minutes)
go run ./main.go rabbitmq load-test --profile heavy

# Sustained load test (10 minutes)
go run ./main.go rabbitmq load-test --profile sustained
```

### Stop RabbitMQ

```bash
# Stop the container
docker-compose down

# Stop and remove volumes (clean reset)
docker-compose down -v
```

## Connection Details

- **Host**: localhost
- **Port**: 5672 (AMQP)
- **Management UI Port**: 15672
- **Username**: guest
- **Password**: guest
- **Virtual Host**: / (default)

## Customizing RabbitMQ Configuration

Edit `docker-compose.yml` to adjust:

### Memory Limit
```yaml
environment:
  RABBITMQ_DEFAULT_USER: guest
  RABBITMQ_DEFAULT_PASS: guest
  RABBITMQ_VM_MEMORY_HIGH_WATERMARK: 512MB
```

### Additional Plugins
```yaml
environment:
  RABBITMQ_PLUGINS: rabbitmq_management rabbitmq_federation
```

### Custom Configuration File
Mount a rabbitmq.conf file:
```yaml
volumes:
  - ./rabbitmq.conf:/etc/rabbitmq/rabbitmq.conf
  - rabbitmq_data:/var/lib/rabbitmq
```

## Useful Commands

```bash
# Check RabbitMQ status
docker-compose exec rabbitmq rabbitmqctl status

# List all queues
docker-compose exec rabbitmq rabbitmqctl list_queues

# List all exchanges
docker-compose exec rabbitmq rabbitmqctl list_exchanges

# List connections
docker-compose exec rabbitmq rabbitmqctl list_connections

# View messages in a queue
docker-compose exec rabbitmq rabbitmqctl list_queues name messages

# Reset RabbitMQ (clear all data)
docker-compose down -v && docker-compose up -d
```

## Load Test Example Commands

```bash
# Quick light load test
docker-compose up -d && \
go run ./main.go rabbitmq load-test --profile light && \
docker-compose down

# Medium load test with custom settings
go run ./main.go rabbitmq load-test \
  --profile medium \
  --duration 3m \
  --parallel-clients 15

# Custom high-throughput test
go run ./main.go rabbitmq load-test \
  --profile heavy \
  --duration 5m \
  --queue-count 100 \
  --exchange-count 50 \
  --message-size 8192 \
  --parallel-clients 32
```

## Troubleshooting

### Container won't start
```bash
# Check logs
docker-compose logs rabbitmq

# Remove and restart
docker-compose down -v
docker-compose up -d
```

### Connection refused
```bash
# Verify container is running
docker-compose ps

# Check if port 5672 is accessible
netstat -an | grep 5672
# or
lsof -i :5672
```

### Memory/CPU issues during tests
Edit `docker-compose.yml` to add resource limits:
```yaml
services:
  rabbitmq:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
```

## Load Test Profiles

| Profile | Duration | Queues | Exchanges | Expected Throughput |
|---------|----------|--------|-----------|-------------------|
| light   | 30s      | 5      | 2         | 2.5K-5K msgs/sec  |
| medium  | 2m       | 20     | 5         | 20K-50K msgs/sec  |
| heavy   | 3m       | 50     | 20        | 100K+ msgs/sec    |
| sustained| 10m     | 30     | 10        | Normal, long-run  |
