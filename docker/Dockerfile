FROM golang:1.17 as builder

ADD . /app
WORKDIR /app/

RUN DEBIAN_FRONTEND=noninteractive apt-get -y install make git

RUN make vendor
RUN make -d

FROM debian:buster-slim as prod

RUN adduser --home /app --disabled-password --gecos "" app ;\
	apt update ;\
	apt install ca-certificates curl -y --no-install-recommends ;\
	rm -rf /var/lib/apt/lists/*

USER app

COPY --from=builder /app/bin/bot /app/bot
COPY --from=builder /app/.env /app/.env


CMD ["/app/bot"]
