FROM rust:latest

# 1. Install system dependencies including networking tools
RUN apt-get update && \
    apt-get install -y \
    libssl-dev \
    clang \
    pkg-config \
    ca-certificates \
    curl \
    iputils-ping \
    dnsutils \
    lshw \
    netcat-openbsd && \
    rm -rf /var/lib/apt/lists/*

# 2. Set up working directory
WORKDIR /app

# 3. Cache dependencies
COPY Cargo.toml Cargo.lock ./

# 4. Copy source code
COPY src ./src
RUN mkdir -p data/embedding_cache/
#COPY data/embedding_cache ./data/embedding_cache

# 5. Build application
RUN cargo build #--release

# 6. Add health check and connection test script
COPY check_connection.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/check_connection.sh

EXPOSE 3030

# 7. Use entrypoint that verifies connectivity first
ENTRYPOINT ["/usr/local/bin/check_connection.sh"]