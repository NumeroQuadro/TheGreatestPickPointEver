FROM golang:1.24.0
RUN curl -fsSL \
        https://raw.githubusercontent.com/pressly/goose/master/install.sh | sh
WORKDIR ./app

COPY go.mod go.sum ./
RUN go mod download

COPY migrations_up.sh /migrations_up.sh
RUN chmod +x /migrations_up.sh
COPY . ./

ENV MIGRATION_FOLDER=./migrations
ENV POSTGRES_SETUP_MAIN="user=test password=test dbname=test host=db port=5432 sslmode=disable"
ENV POSTGRES_SETUP_TEST="user=test password=test dbname=test host=db port=5433 sslmode=disable"

#RUN goose -dir ${MIGRATION_FOLDER} postgres "${POSTGRES_SETUP_MAIN}" up

RUN CGO_ENABLED=0 GOOS=linux go build -o /hw-4 ./cmd/main.go

EXPOSE 9000

ENTRYPOINT ["/migrations_up.sh"]
