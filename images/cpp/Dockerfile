FROM debian:stable-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    g++ \
    && rm -rf /var/lib/apt/lists/* \
    libstdc++-dev \
    && useradd -m -u 1000 runner
USER runner
WORKDIR /app
CMD ["tail", "-f", "/dev/null"]