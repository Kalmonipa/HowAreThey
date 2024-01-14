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
RUN CGO_ENABLED=0 GOOS=linux go build -o howarethey

# Second stage: create the runtime image
FROM alpine:3.19

# Set the working directory in the new image
WORKDIR /root/

# Copy the pre-built binary file from the first stage
COPY --from=builder /app/howarethey .

# Copy the config files
COPY config/friends.yaml config/friends.yaml

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./howarethey"]
