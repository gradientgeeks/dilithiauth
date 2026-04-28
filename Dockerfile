# Stage 1: Build React frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binaries
FROM golang:1.23-alpine AS go-builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy compiled React assets into the embed path before building
COPY --from=frontend-builder /src/cmd/server/static ./cmd/server/static
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server     ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/management ./cmd/management

# Stage 3: Minimal runtime image
FROM gcr.io/distroless/static-debian12
COPY --from=go-builder /bin/server     /server
COPY --from=go-builder /bin/management /management
VOLUME ["/data"]
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
