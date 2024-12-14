# BiTaksi Case Study

This project contains a microservice architecture that manages driver locations and matches passengers with the nearest drivers.

## Services

1. **Driver Location API** (Port 8080)
   - Manages driver locations
   - Processes location updates
   - Provides functionality to find nearby drivers

2. **Matching API** (Port 8081)
   - Matches passengers with the nearest drivers

## Installation

1. Clone the project:
```bash
git clone https://github.com/yusufatac/bitaksi-case-study.git
cd bitaksi-case-study
```

2. Set up environment variables:
```bash
cp deployments/.env_dist deployments/.env
```

3. Run with Docker Compose:
```bash
cd deployments
docker-compose up --build
```

## API Endpoints

### Authentication

#### Register - POST /api/v1/auth/register
```json
{
  "username": "string",
  "password": "string",
  "email": "string"
}
```

#### Login - POST /api/v1/auth/login
```json
{
  "username": "string",
  "password": "string"
}
```

### Driver Location API

#### Update Location - POST /api/v1/locations
```json
{
  "driver_id": "string",
  "latitude": 0.0,
  "longitude": 0.0
}
```

#### Batch Update Locations - POST /api/v1/locations/batch
```json
[
  {
    "driver_id": "string",
    "latitude": 0.0,
    "longitude": 0.0
  }
]
```

#### Find Nearby Drivers - POST /api/v1/locations/nearby
```json
{
  "latitude": 0.0,
  "longitude": 0.0,
  "radius": 0.0
}
```

### Matching API

#### Find Nearest Driver - POST /api/v1/match
```json
{
  "latitude": 0.0,
  "longitude": 0.0,
  "radius": 0.0
}
```

## Monitoring

Both services provide health checks through the `/health` endpoint.

## License

MIT