# Use a minimal image to run the Go app
FROM alpine:3.21

ARG KIND

# Install ffmpeg
RUN apk add --no-cache ffmpeg opus ca-certificates libgcc libstdc++

# Copy the Pre-built binary file from the previous stage
COPY ${KIND} app

RUN chmod +x app

# Command to run the executable
CMD ./app