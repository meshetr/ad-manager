FROM golang:1.15-alpine AS build
WORKDIR /src
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app .

FROM scratch
COPY --from=build /app /
EXPOSE 8080
ENTRYPOINT ["/app"]