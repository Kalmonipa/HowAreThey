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

# Create a group and user
# Replace '1000' with the desired GID if different
RUN groupadd -g 1000 hat && \
    useradd -r -u 1000 -g hat hat

# Set the working directory in the new image
WORKDIR /home/hat

# Install any necessary dependencies
RUN apt-get update && apt-get install --assume-yes --no-install-recommends \
    libsqlite3-0=3.34.1-3 \
    ca-certificates=20210119 \
    && rm -rf /var/lib/apt/lists/*

# Copy the pre-built binary file from the first stage
COPY --from=builder /app/howarethey .

# Change ownership of the working directory
# This step ensures that the 'hat' user has the necessary permissions
RUN mkdir sql \
    && chown -R hat:hat /home/hat

# Create the sql directory

# Expose the port the app runs on
EXPOSE 8080

# Use the created user to run the app
USER hat

# Command to run the executable
CMD ["./howarethey"]
