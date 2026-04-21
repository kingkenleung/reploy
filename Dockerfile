FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY reploy-linux .
RUN mv reploy-linux reploy
COPY web/ web/
EXPOSE 3000
CMD ["./reploy"]
