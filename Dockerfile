FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm install

COPY frontend/ ./
RUN npx vite build --outDir dist

FROM golang:1.22-alpine AS backend-builder
WORKDIR /app/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

RUN mkdir -p cmd/api/dist
COPY --from=frontend-builder /app/frontend/dist/ ./cmd/api/dist/

RUN CGO_ENABLED=0 go build -o /api-server ./cmd/api/main.go

FROM alpine:latest
WORKDIR /app

RUN mkdir -p /app/data

COPY --from=backend-builder /api-server /app/api-server

CMD ["/app/api-server"]
