# Feed-Pulse-Back

## Overview

Feed-Pulse-Back is the backend service for the Feed-Pulse project, a web application designed to analyze and visualize feedback data. This service is maked with golang and [fiber](https://gofiber.io/) framework.


## Features

- User authentication and authorization
- Feedback data management
- Data analysis and visualization
- RESTful API for frontend integration


## Technologies Used

- Go (Golang)
- Fiber framework
- PostgreSQL (or any other database of your choice)
- JWT for authentication
- Docker for containerization
- Swagger for API documentation

## Getting Started

### Prerequisites

- GO 1.24 or higher
- PostgreSQL (or any other database of your choice)
- GCC (needed for some Go packages)
- Docker (optional, for containerization)
- Mistral API key [https://mistral.ai/](https://mistral.ai/)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/ynov-2025-m1-team6/Feed-Pulse-Back
   ```
2. Navigate to the project directory:
   ```bash
    cd Feed-Pulse-Back
    ```
3. Install dependencies:
    ```bash
    go mod tidy
    ```
4. Set up the database:
    - Create a PostgreSQL database and user.
    - Update the database connection string in the `.env` file.

5. Add your Mistral API key in the `.env` file:
    ```bash
    MISTRAL_API_KEY=your_mistral_api_key
    ```
6. Run the application:
    ```bash
    go run cmd/app/main.go
    ```

7. Access the API documentation at `http://localhost:3000/swagger/index.html`.

8. Use the API endpoints to interact with the application.

## Environment Deployment

### Development

Url: `https://feed-pulse-api-dev.onrender.com`

- The development environment is hosted on Render and is automatically deployed from the `dev` branch of the repository.

### Production

Url: `https://feed-pulse-api.onrender.com`

- The production environment is hosted on Render and is automatically deployed from the `main` branch of the repository.

## Authors

- [Tot0p](https://github.com/tot0p)
- [Mkarten](https://github.com/mkarten)
- [Axou89](https://github.com/Axou89)