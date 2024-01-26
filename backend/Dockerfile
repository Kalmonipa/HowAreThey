# First stage: build the application
# Use an official Go image as the base image
FROM golang:1.18 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./

# Download necessary Go modules
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o howarethey

# Second stage: create the runtime image
# Use a slim Debian or Ubuntu base image
FROM debian:bullseye-slim

# Set the working directory in the new image
WORKDIR /root/

# Install any necessary dependencies
RUN apt-get update && apt-get install --assume-yes --no-install-recommends \
# https://pkgs.org/download/libsqlite3-0
    libsqlite3-0=3.34.1-3 \
    ca-certificates=20210119 \
    && rm -rf /var/lib/apt/lists/*

# Copy the pre-built binary file from the first stage
COPY --from=builder /app/howarethey .

# Copy the config files
COPY sql/friends.db sql/friends.db

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./howarethey"]
