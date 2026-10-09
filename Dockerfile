FROM node:22-alpine AS frontend-build

WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --legacy-peer-deps
COPY frontend/ ./
RUN npm run build

FROM golang:1.22-alpine AS backend-build

WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN go test ./... \
    && mkdir -p /out \
    && CGO_ENABLED=0 GOOS=linux go build -o /out/devboard .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

COPY --from=backend-build /out/devboard /usr/local/bin/devboard
COPY --from=frontend-build /src/frontend/dist /app/frontend

ENV FRONTEND_DIR=/app/frontend
EXPOSE 8080

USER app
ENTRYPOINT ["/usr/local/bin/devboard"]
