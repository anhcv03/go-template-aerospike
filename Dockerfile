ARG BUILDER_IMAGE=harbor.vht.vn/c4i/golang:1.24.0-bookworm
ARG TARGET_IMAGE=scratch

FROM ${BUILDER_IMAGE} AS builder

WORKDIR /utm-track-manager

COPY . .

RUN make go-build


FROM ${TARGET_IMAGE} AS final

WORKDIR /utm-track-manager

COPY --from=builder /utm-track-manager/bin/utm-track-manager ./bin/utm-track-manager
COPY --from=builder /utm-track-manager/web ./bin
COPY --from=builder /utm-track-manager/app.yaml ./bin/app.yaml
COPY --from=builder /utm-track-manager/data ./data

ENTRYPOINT ["./bin/utm-track-manager", "start"]
CMD ["./bin"]