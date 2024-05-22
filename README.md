# Notification system

A nofication-sending system written in Go. It covers the following requirements:

* Send notification over email, SMS, Slack (and extensible for more)
* Scale horizontally
* Guarantee an "at least once" SLA for sending a notification
* Provide HTTP API for receiving a notification request event

### Design

The system consists of four components.

* API service - accepts API requests and writes the messages into Kafka topics
* Kafka - persists the messages
* Distributor services - read the messages from Kafka topics and send them to email, SMS or Slack
* Redis - persists the current offset for each topic partition

For message queue I chose Kafka because of its fault-tolerance and scalability.

For key-value store I chose Redis because of its high-performance with regards to quick retrieval and modification of data.

### Workflow

At endpoint `/api/v1/notification` the API service accepts messages that contain the notification channel(`email`, `sms` or `slack`) and the text of the notification that needs to be sent.

The service parses the message and writes it into a corresponding Kafka topic. For ease of use, automatic topic creation is set to true.

The distributor services(at least one container for each notification channel) have a topic consumer for a specified Kafka topic. While starting or rebalancing a consumer, the service queries Redis and sets exact offset to each topic partition.

When a message is consumed, the service delivers the message to the corresponding notification channel and saves the current offset in Redis. This ensures "at least once" delivery by putting only successfully delivered message offsets in Redis.

### Deploy in Docker
Build the Docker images.
```
cd api && docker build -t api:1 . && cd ..
cd distributor && docker build -t distributor:1 . && cd ..
```
Open `docker-compose.yml` and in the distributor services fill out the empty environment variables. If you want to test the SMS-sending please sign up to `twilio.com` to get the required values.

When you are ready you can start the containers.
```
DIR=$(pwd) docker compose up -d
```
Docker compose will create directory `data` that will contain the data stored by Kafka and Redis.

<em>Note: Docker compose will bring up all of the containers at the same time. Kafka takes some seconds to initiate. During that time the api and distributor services will try to connect periodically. </em>

To discard the deployment delete the containers manually or use docker compose again.
```
docker compose down
```

### Example API requests
```
curl -i -X POST -H "Content-Type: application/json" -d '{"message": "Charm!", "channel": "email"}' http://localhost:8080/api/v1/notification

curl -i -X POST -H "Content-Type: application/json" -d '{"message": "Charm!", "channel": "sms"}' http://localhost:8080/api/v1/notification

curl -i -X POST -H "Content-Type: application/json" -d '{"message": "Charm!", "channel": "slack"}' http://localhost:8080/api/v1/notification
```

### Scalability and production deployment 

We are going to deploy our system in a Kubernetes cluster. How will the system scale?

* For the API service we can create a HorizontalPodAutoscaler that scales the pods based on CPU utilization.

* In order to scale the distributor services we need to monitor the lag in the Kafka topics. We can set up Prometheus for lag metric collection and an [adapter](https://github.com/kubernetes-sigs/prometheus-adapter) that exposes the metrics. Then we can create a HorizontalPodAutoscaler that scales the pods based on the custom metrics we've exported with Prometheus.

* Kafka and Redis should be scaled according to their documentation.