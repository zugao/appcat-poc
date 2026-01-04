FROM docker.io/library/alpine:3.15

ENTRYPOINT ["function-appcat-poc"]

# Install runtime dependencies
RUN apk add --update --no-cache \
    bash \
    ca-certificates \
    curl

# Create cache directory with proper permissions
RUN mkdir /.cache && chmod -R g=u /.cache

# Copy the binary
COPY function-appcat-poc /usr/local/bin/
COPY package.yaml /package.yaml

# Make binary executable
RUN chmod a+x /usr/local/bin/function-appcat-poc

# Run as non-root user
USER 65532:0
