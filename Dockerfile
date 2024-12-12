# Use the official Go image as a base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy wait-for-it.sh
COPY wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

# RUN go install github.com/cosmtrek/air@latest

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the Go application
RUN go build -o main ./cmd/oms-api/main.go

# Expose port 8080 (or any other port your app runs on)
EXPOSE 8080

# Command to run the application
# CMD ["./main"]
CMD ["/wait-for-it.sh", "postgres-container-33:5432", "--", "./main"]


