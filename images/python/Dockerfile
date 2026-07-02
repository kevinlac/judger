FROM python:3.12-slim
RUN useradd -m -u 1000 runner
USER runner
WORKDIR /app
CMD ["tail", "-f", "/dev/null"]