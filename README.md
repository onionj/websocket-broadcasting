
# Websocket Broadcasting with Go-Fiber and RabbitMQ

![websocket.gif](./docs/websocket.gif)

This project demonstrates a WebSocket broadcasting application with multi-processing capabilities using Fiber and RabbitMQ as the message broker. It allows multiple processes to handle WebSocket connections and efficiently broadcast messages to connected clients.

Please note that the project uses the following technologies:

- [Go-Fiber](https://github.com/gofiber/fiber): A fast and efficient web framework for Go.
- [RabbitMQ](https://www.rabbitmq.com/): A powerful and flexible open-source message broker.

## Prerequisites

Before running the application, ensure you have the following installed:

- Go (1.16 or higher): https://golang.org/dl/
- RabbitMQ Server: https://www.rabbitmq.com/download.html

## Installation

1. Clone this repository to your local machine:

```
git clone https://github.com/onionj/websocket-broadcasting.git
cd websocket-broadcasting
```

2. Fetch the required dependencies using Go Modules:

```
go mod download
```

## Configuration

Ensure you have RabbitMQ running on your local machine or update the configuration settings in `config.json` to match your RabbitMQ server.

## Usage

To run the WebSocket broadcasting application, use the following command:

```
go run main.go
```

The application will start, and you can visit `http://localhost:8080` in your browser to interact with the WebSocket server.

## How it Works

The application uses Fiber to handle incoming WebSocket connections and messages. When a WebSocket connects, the server add it to a list.

Each process has a queue bound to a FANOUT exchange. When a message is pushed to the exchange, all queues receive the message, and each process sends the message to the connected websockets.
