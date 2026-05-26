# syntax=docker/dockerfile:1.7

# ---------- Stage 1: build app binaries + isolate sandbox ----------
FROM golang:1.25-bookworm AS builder

# Build deps for isolate (asciidoc is needed because upstream `make` builds the manpage).
RUN apt-get update && apt-get install -y --no-install-recommends \
        git ca-certificates build-essential pkg-config \
        asciidoc libxml2-utils xsltproc docbook-xml docbook-xsl \
        libcap-dev libseccomp-dev libsystemd-dev \
 && rm -rf /var/lib/apt/lists/*

# Clone isolate, run its own Makefile.
ARG ISOLATE_REF=master
RUN git clone --depth=1 --branch ${ISOLATE_REF} https://github.com/ioi/isolate.git /tmp/isolate \
 && make -C /tmp/isolate \
 && make -C /tmp/isolate install DESTDIR=/install

WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Compile the three binaries.
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
 && CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker \
 && CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/seed   ./cmd/seed


# ---------- Stage 2: runtime image ----------
FROM ubuntu:24.04 AS runtime

ENV DEBIAN_FRONTEND=noninteractive \
    PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

# Language toolchains + isolate's runtime libs + shadow utils (for useradd).
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates curl passwd \
        libcap2 libcap2-bin libseccomp2 libsystemd0 \
        python3.12 \
        g++-13 \
 && update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-13 100 \
 && update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-13 100 \
 && ln -sf /usr/bin/python3.12 /usr/local/bin/python3 \
 && ln -sf /usr/bin/python3.12 /usr/local/bin/python \
 && rm -rf /var/lib/apt/lists/*

# Go 1.22 for user submissions (separate from the 1.25 build toolchain in stage 1).
ARG GO_SUBMISSION_VERSION=1.22.12
RUN curl -fsSL "https://go.dev/dl/go${GO_SUBMISSION_VERSION}.linux-amd64.tar.gz" \
        | tar -C /usr/local -xz \
 && ln -sf /usr/local/go/bin/go /usr/local/bin/go \
 && ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt

# Pull the entire `make install` tree (binary, helper, cg-keeper, default.cf,
# manpage, sysusers.d snippet) in one COPY — no per-file picking, no sed.
COPY --from=builder /install/ /

# Create the `isolate` system user that the upstream-shipped default.cf expects,
# with auto-allocated subuid/subgid (shadow >= 4.9 supports --add-subids-for-system).
RUN useradd --system --user-group --no-create-home --shell /usr/sbin/nologin \
            --add-subids-for-system isolate \
 && mkdir -p /var/local/lib/isolate /run/isolate

COPY --from=builder /out/server /usr/local/bin/executify-server
COPY --from=builder /out/worker /usr/local/bin/executify-worker
COPY --from=builder /out/seed   /usr/local/bin/executify-seed

WORKDIR /app
COPY migrations ./migrations
COPY seed       ./seed
COPY configs    ./configs

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/executify-server"]
