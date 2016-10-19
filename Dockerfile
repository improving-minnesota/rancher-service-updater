# Docker image for the Rancher Service Update
#
#     docker build --rm=true -t rancher-service-updater

FROM gliderlabs/alpine:3.2
RUN apk add --update \
  ca-certificates
ADD rancher-service-updater /bin/
ENTRYPOINT ["/bin/rancher-service-updater"]
