# Stage 1: Modules caching
FROM golang:1.24 as modules
COPY go.mod go.sum ./
RUN go mod download

# Stage 2: Build
FROM golang:1.24 as builder
WORKDIR /app
COPY --from=modules /go/pkg /go/pkg
COPY . .

# Install playwright CLI
RUN PWGO_VER=$(grep -oE "playwright-go v\S+" go.mod | sed 's/playwright-go //g') \
    && go install github.com/playwright-community/playwright-go/cmd/playwright@${PWGO_VER}
# Build your app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o screenshoter ./cmd/main.go


# Stage 3: Final
FROM ubuntu:noble

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# Copy binaries
COPY --from=builder /go/bin/playwright /usr/local/bin/playwright
COPY --from=builder /app/screenshoter /app/screenshoter

# Install Playwright browsers and dependencies
RUN apt-get update && apt-get install -y \
    wget \
    && playwright install --with-deps \
    && apt-get remove -y wget \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
CMD ["/app/screenshoter"]