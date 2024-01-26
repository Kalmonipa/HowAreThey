# First stage: Build the Go application
FROM golang:1.18 as go-builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/. .
RUN CGO_ENABLED=1 GOOS=linux go build -o howarethey

# Second stage: Build the React (Node.js) application
FROM node:14 as node-builder
WORKDIR /usr/src/app
COPY frontend/. .
RUN npm install && npm run build

# Third stage: Combine both in a runtime image
FROM debian:bullseye-slim
WORKDIR /root/

# Install Node.js and other dependencies
RUN apt-get update && apt-get install --assume-yes --no-install-recommends \
    nodejs=12.22.12~dfsg-1~deb11u4 \
    npm=7.5.2+ds-2 \
    libsqlite3-0=3.34.1-3 \
    ca-certificates=20210119 \
    && rm -rf /var/lib/apt/lists/*

# Copy the Go application
COPY --from=go-builder /app/howarethey .

# Copy the React application
COPY --from=node-builder /usr/src/app /usr/src/app

COPY sql/ ./sql

# Expose ports for both applications
EXPOSE 3000 8080

# Run both applications
CMD ["sh", "-c", "./howarethey & cd /usr/src/app && npm run start"]
