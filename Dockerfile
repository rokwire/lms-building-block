FROM golang:1.16-buster as builder

ENV CGO_ENABLED=0

RUN mkdir /lms-app
WORKDIR /lms-app
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
RUN make

FROM alpine:3.13

#we need timezone database
RUN apk --no-cache add tzdata

COPY --from=builder /lms-app/bin/lms /
COPY --from=builder /lms-app/docs/swagger.yaml /docs/swagger.yaml

COPY --from=builder /lms-app/driver/web/authorization_model.conf /driver/web/authorization_model.conf
COPY --from=builder /lms-app/driver/web/authorization_policy.csv /driver/web/authorization_policy.csv

#we need timezone database
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo 

ENTRYPOINT ["/lms"]
