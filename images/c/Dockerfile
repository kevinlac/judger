FROM debian:stable-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/* \
    && useradd -m -u 1000 runner
USER runner
WORKDIR /app
CMD ["tail", "-f", "/dev/null"]