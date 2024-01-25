FROM golang:1.21-alpine as builder

ENV CGO_ENABLED=0

RUN apk add --no-cache --update make git

RUN mkdir /lms-app
WORKDIR /lms-app
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
RUN make

FROM alpine:3.17

#we need timezone database
RUN apk --no-cache add tzdata

COPY --from=builder /lms-app/bin/lms /
COPY --from=builder /lms-app/driver/web/docs/gen/def.yaml /driver/web/docs/gen/def.yaml

COPY --from=builder /lms-app/driver/web/admin_permission_policy.csv /driver/web/admin_permission_policy.csv
COPY --from=builder /lms-app/driver/web/client_permission_policy.csv /driver/web/client_permission_policy.csv
COPY --from=builder /lms-app/driver/web/client_scope_policy.csv /driver/web/client_scope_policy.csv

COPY --from=builder /lms-app/vendor/github.com/rokwire/core-auth-library-go/v2/authorization/authorization_model_string.conf /lms-app/vendor/github.com/rokwire/core-auth-library-go/v2/authorization/authorization_model_string.conf


#we need timezone database
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo 

ENTRYPOINT ["/lms"]
