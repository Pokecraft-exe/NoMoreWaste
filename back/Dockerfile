# Back-end NO MORE WASTE (oauth + api fusionnes, un seul processus, port
# 8080, TLS). A construire depuis ce dossier :
#   docker build -t nomorewaste-backend .
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/backend .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates openssl \
    && addgroup -S app && adduser -S -G app app \
    && mkdir -p /app/DATA /certs \
    && openssl req -x509 -nodes -days 825 -newkey rsa:2048 \
        -keyout /certs/api.key -out /certs/api.crt \
        -subj "/C=FR/O=NoMoreWaste/CN=nomorewaste-backend.local" \
        -addext "basicConstraints=critical,CA:FALSE" \
        -addext "keyUsage=digitalSignature,keyEncipherment" \
        -addext "subjectAltName=DNS:nomorewaste-backend.local,DNS:localhost,IP:127.0.0.1,IP:192.168.100.128" \
    && chown -R app:app /app /certs
WORKDIR /app
COPY --from=build /out/backend ./backend
USER app

ENV PORT=8080 \
    DATA_DIR=/app/DATA \
    TLS_CERT_FILE=/certs/api.crt \
    TLS_KEY_FILE=/certs/api.key
EXPOSE 8080

ENTRYPOINT ["./backend"]
