ARG COMMAND
FROM ${COMMAND}-app as app
# This is where one could build the application code as well.

FROM public.ecr.aws/lambda/provided:al2
# Copy binary to production image.
COPY --chmod=0755 --from=app /app/main /var/runtime/main
COPY --chmod=0755 bootstrap.sh /var/runtime/bootstrap

# Copy Tailscale binaries from the tailscale image on Docker Hub.
COPY --from=docker.io/tailscale/tailscale:stable /usr/local/bin/tailscaled /var/runtime/tailscaled
COPY --from=docker.io/tailscale/tailscale:stable /usr/local/bin/tailscale /var/runtime/tailscale
RUN mkdir -p /var/run && ln -s /tmp/tailscale /var/run/tailscale && \
    mkdir -p /var/cache && ln -s /tmp/tailscale /var/cache/tailscale && \
    mkdir -p /var/lib && ln -s /tmp/tailscale /var/lib/tailscale && \
    mkdir -p /var/task && ln -s /tmp/tailscale /var/task/tailscale

# Run on container startup.
ENTRYPOINT ["/var/runtime/bootstrap"]