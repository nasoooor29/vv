# Monitoring & Health

Visory provides metrics and health monitoring to track system and service status.

## Overview

The monitoring system provides:
- System health status
- Service-level metrics
- Error rate tracking
- Real-time health checks

## Required Permissions

| Action | Permission |
|--------|------------|
| View metrics | `audit_log_viewer` |
| View health | `health_checker` |

## Health Monitoring

### System Health

The health endpoint provides overall system status:

```
GET /api/health
```

Response:
```json
{
  "database": "healthy"
}
```

### Service Health

Detailed per-service health status:

```
GET /api/metrics/health
```

Response:
```json
{
  "status": "healthy",
  "services": {
    "auth": {
      "status": "healthy",
      "error_rate": 0.02,
      "error_count": 5,
      "total_count": 250
    },
    "docker": {
      "status": "warning",
      "error_rate": 0.08,
      "error_count": 20,
      "total_count": 250
    }
  },
  "alerts": [
    {
      "service": "docker",
      "message": "Error rate above 5% threshold"
    }
  ]
}
```

### Health Thresholds

| Status | Condition |
|--------|-----------|
| Healthy | Error rate < 5% |
| Warning | Error rate 5-10% |
| Critical | Error rate > 10% |

## Metrics Dashboard

Access the monitoring dashboard at **Monitor** in the sidebar.

### Available Metrics

| Metric | Description |
|--------|-------------|
| Error Rate | Percentage of failed operations by service |
| Log Count | Number of log entries over time |
| Level Distribution | Breakdown by log level |
| Service Distribution | Breakdown by service |

## API Reference

### Get All Metrics

```
GET /api/metrics
```

Response:
```json
{
  "error_rate_by_service": {
    "auth": 0.02,
    "docker": 0.08,
    "qemu": 0.01
  },
  "log_count_by_hour": [
    {"hour": "2024-01-15T10:00:00Z", "count": 45},
    {"hour": "2024-01-15T11:00:00Z", "count": 62}
  ],
  "level_distribution": {
    "DEBUG": 100,
    "INFO": 850,
    "WARN": 40,
    "ERROR": 10
  },
  "service_distribution": {
    "auth": 300,
    "docker": 400,
    "qemu": 200,
    "users": 100
  },
  "analysis_period": {
    "start": "2024-01-08T00:00:00Z",
    "end": "2024-01-15T23:59:59Z"
  }
}
```

### Get Service-Specific Metrics

```
GET /api/metrics/:service
```

Returns metrics filtered to a specific service.

## WebSocket Real-Time Updates

Connect to the WebSocket endpoint for real-time updates:

```
GET /api/websocket
```

The server sends a heartbeat with timestamp every 2 seconds:

```json
{
  "type": "heartbeat",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Frontend Health Store

The frontend maintains a health store that polls the health endpoint:

```typescript
import { useHealthStore } from '@/stores/health';

function HealthIndicator() {
  const { database } = useHealthStore();
  
  return (
    <span className={database === 'healthy' ? 'text-green-500' : 'text-red-500'}>
      Database: {database}
    </span>
  );
}
```

## Alerting

Services with high error rates generate alerts:

| Threshold | Action |
|-----------|--------|
| > 5% | Warning alert |
| > 10% | Critical alert |

Alerts include:
- Service name
- Error rate
- Alert message

## Best Practices

1. **Monitor Regularly**: Check the dashboard for anomalies
2. **Set Up Alerts**: Configure notifications for critical issues
3. **Track Trends**: Watch for increasing error rates over time
4. **Investigate Errors**: Drill down into error logs when rates spike
5. **Capacity Planning**: Use metrics to plan scaling and resources
