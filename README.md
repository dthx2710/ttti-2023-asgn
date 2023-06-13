# Tiktok Tech Immersion 2023 Backend Server Assignment

This repository contains the source code for the Tiktok Tech Immersion Backend Server Assignment project.

The project is a scalable messaging application backend that consists of an HTTP server, an RPC server, both coded in Go.

Data is stored in a Redis database.

## Links
- [TikTok Tech Immersion 2023](https://bytedance.sg.feishu.cn/docx/CEusdOSGHody93xCekHlbBOvgGR)
- [Project Description](https://bytedance.sg.feishu.cn/docx/P9kQdDkh5oqG37xVm5slN1Mrgle)
- [Code Template Repository](https://github.com/TikTokTechImmersion/assignment_demo_2023)
- [Project Walkthrough](https://o386706e92.larksuite.com/docx/QE9qdhCmsoiieAx6gWEuRxvWsRc) by [weixingp](https://www.linkedin.com/in/weixingp/)

## Directories

- `http-server`: Contains the source code for the HTTP server component.
- `rpc-server`: Contains the source code for the RPC server component.

## Prerequisites

Before running the project, make sure you have the following dependencies installed:

- Docker
- Docker Compose

## Getting Started

To start the project using Docker Compose, follow these steps:

1. Clone this repository to your local machine:
`git clone https://github.com/dthx2710/ttti-2023-asgn.git`

2. Navigate to the project directory:
`cd ttti-2023-asgn`


3. Build the Docker images for the HTTP server, RPC server, and web application:
`docker-compose build`


4. Start the containers:
`docker-compose up`


This will start the HTTP server, RPC server, Redis database, and they will be accessible on the specified ports.

## Usage

Once the containers are up and running, you can access the components using the following URLs:

- HTTP Server: http://localhost:8080
- RPC Server: http://localhost:8888

API Endpoints are defined in the `http-server/main.go` file.
- `/api/send` - POST
- 
Example body:
```
{
    "sender": "user1",
    "receiver": "user2",
    "text": "hello"
}
```

- `/api/pull` - GET
Example body:
```
{
    "chat": "user1:user2",
    "cursor": 0
    "limit": 10
    "reverse": true
}
```

The recommended way is to use Postman to send HTTP requests to the HTTP server.

Feel free to explore the project directories and make any necessary modifications to fit your requirements.


## Credits
Credits to [weixingp](https://www.linkedin.com/in/weixingp/) for the project walkthrough (especially on Redis for rpc-server handlers setup), and TikTok for code template & the immersion course.