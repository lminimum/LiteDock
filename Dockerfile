FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

FROM golang:1.25-alpine3.21 AS go-builder

COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

COPY . /app
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags migrate -o /bin/app ./cmd/app

FROM nginx:1.31.3-alpine

COPY --from=go-builder /bin/app /app
COPY --from=go-builder /app/config /config
COPY --from=go-builder /app/migrations /migrations

COPY --from=frontend-builder /app/dist /usr/share/nginx/html
COPY nginx/docker-nginx.conf /etc/nginx/nginx.conf

EXPOSE 80

CMD /app & exec nginx -g "daemon off;"
