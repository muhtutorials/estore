FROM golang

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

ENV PROJECT_DIR=/app GO111MODULE=on CGO_ENABLED=0

RUN go get github.com/githubnemo/CompileDaemon

RUN go install github.com/githubnemo/CompileDaemon

CMD ["CompileDaemon", "-build=go build -o /build/app", "-command=/build/app", "-polling=true", "-polling-interval=2000"]