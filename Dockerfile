FROM node:24-bookworm-slim AS frontend
WORKDIR /src
RUN corepack enable && corepack prepare pnpm@11.28.13 --activate
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web ./web
RUN pnpm run build

FROM golang:1.27.4-bookworm AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY --from=frontend /src/internal/webui/assets ./internal/webui/assets
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /pnd ./cmd/pnd

FROM scratch
COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=backend /pnd /pnd
COPY LICENSE THIRD_PARTY_NOTICES.md /licenses/
COPY third_party/licenses /licenses/dependencies
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/pnd"]
