# STEP 1 build executable binary
FROM alpine as builder
# Create nonroot user
RUN adduser -D -g '' registry-cleanup-user
# Add ca-certificates
RUN apk --update add ca-certificates

# STEP 2 build a small image from scratch
FROM scratch
LABEL maintainer="cblomart@gmail.com"
# copy password file for users
COPY --from=builder /etc/passwd /etc/passwd
# copy ca-certificates 
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# copy registry-cleanup
COPY ./registry-cleanup /registry-cleanup
# run as registry-cleanup-user
USER registry-cleanup-user
# start registry-cleanup
ENTRYPOINT [ "/registry-cleanup" ] 