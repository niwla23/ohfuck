# ------------------ #
# build the frontend #
# ------------------ #

FROM node:18 as frontend_builder
WORKDIR /build
RUN npm i -g pnpm@7.22
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
ADD frontend .
RUN pnpm build

# ------------------ #
# build the backend  #
# ------------------ #

FROM golang:1.19-alpine AS backend_builder
# RUN apk add --no-cache build-base
WORKDIR /tmp/ohfuck
# first only copy files relevant for module downloads (caching)
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
COPY --from=frontend_builder /build/dist /tmp/ohfuck/fake_frontend
# Unit tests
# RUN CGO_ENABLED=0 go test -v
RUN go build -o ./ohfuck .

# ----------- #
# FINAL IMAGE #
# ----------- # 

FROM alpine
# RUN apk add ca-certificates
COPY --from=backend_builder /tmp/ohfuck/ohfuck /app/ohfuck
EXPOSE 3000
CMD ["/app/ohfuck"]