# IOT - Cloud

IOT Cloud is a collection of Golang microservices that operate as event-driven systems, designed to work seamlessly with the IOT Agent to command, monitor, and track devices.

Pre-Alpha Discussions: [Here](https://github.com/BenMeehan/iot-cloud/discussions/5)

## Services

### Heartbeat Service
This service receives device heartbeats through MQTT and stores them in TimescaleDB.

### Registration Service
This service listens for device registration requests over MQTT, validates device secrets, generates unique device IDs, and stores device information in a PostgreSQL (TimescaleDB) database.

### Metrics Service
This service collects and processes system and process metrics from IoT devices via MQTT. It stores the collected metrics in TimescaleDB for monitoring and analysis.

## Running the Project
To run the project, execute:
```bash
go run cmd/main.go
```

## To Add a New Service
1. Create a new folder at the root of the project and add your service there.

## Useful URLs

- [Public MQTT Broker](https://www.emqx.com/en/mqtt/public-mqtt5-broker)
- [TimescaleDB](https://www.timescaledb.com)
