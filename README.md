# IOT - Cloud

IOT cloud a bunch of golang microservices event-driven microservices which can work in tandem with the IOT Agent to command, monitor and track devices. 

## Heartbeat Service
This service recieves device heartbeats through MQTT and stores it in timescale DB.

## Running the project
To run, do
```
go run cmd/main.go
```
## To add a new service
1. Create a new folder at the root of the project and add your service there.

### URLS

https://www.emqx.com/en/mqtt/public-mqtt5-broker
https://www.timescaledb.com
