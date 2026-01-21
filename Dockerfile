FROM public.ecr.aws/docker/library/golang:1.24-alpine as builder

ENV CGO_ENABLED=0

RUN apk add --no-cache --update make git

RUN mkdir /lms-app
WORKDIR /lms-app
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
RUN make

FROM public.ecr.aws/docker/library/alpine:3.21.3

#we need timezone database + certificates
RUN apk add --no-cache tzdata ca-certificates

COPY --from=builder /lms-app/bin/lms /
COPY --from=builder /lms-app/driver/web/docs/gen/def.yaml /driver/web/docs/gen/def.yaml

COPY --from=builder /lms-app/driver/web/admin_permission_policy.csv /driver/web/admin_permission_policy.csv
COPY --from=builder /lms-app/driver/web/client_permission_policy.csv /driver/web/client_permission_policy.csv
COPY --from=builder /lms-app/driver/web/client_scope_policy.csv /driver/web/client_scope_policy.csv

COPY --from=builder /lms-app/vendor/github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/authorization/authorization_model_scope.conf /lms-app/vendor/github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/authorization/authorization_model_scope.conf
COPY --from=builder /lms-app/vendor/github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/authorization/authorization_model_string.conf /lms-app/vendor/github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/authorization/authorization_model_string.conf

ENTRYPOINT ["/lms"]
