#
# Docker NodeJS TypeScript Production-Ready Configuration
#
ARG NODE_VERSION=24.2.0
ARG ALPINE_VERSION=3.21

# Build stage - Compile TypeScript and install dependencies
FROM node:${NODE_VERSION}-alpine${ALPINE_VERSION} AS build

# Install security updates and build dependencies
RUN apk update && apk upgrade && \
    apk add --no-cache --virtual .build-deps \
    make \
    gcc \
    g++ \
    python3 \
    && rm -rf /var/cache/apk/*

# Set working directory to App dir
WORKDIR /app

# Copy package files first for better caching
COPY package*.json ./
COPY tsconfig.json ./

# Install dependencies (including dev dependencies for build)
RUN npm ci --only=production=false && \
    npm cache clean --force

# Copy source code
COPY . .

# Create environment file if it doesn't exist
RUN [ -f .env ] || cp .env.example .env || echo "# Environment variables" > .env

# Install dependencies
RUN npm install

FROM node:24.2.0-alpine3.21 as app

## Copy built node modules and binaries without including the toolchain
COPY --from=build /app .

WORKDIR /app

CMD [ "/app/scripts/run.sh" ]

# Remove dev dependencies and clean up
RUN #npm prune --production && \
    npm cache clean --force && \
    apk del .build-deps
