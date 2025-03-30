# Please read the following links before build:
# https://github.com/grafana/k6
# https://github.com/grafana/xk6
# https://github.com/phymbert/xk6-sse
# https://hub.docker.com/r/grafana/xk6/

docker run --rm -it -e GOOS=darwin -u "$(id -u):$(id -g)" -v "${PWD}:/xk6" \
  grafana/xk6 build v0.58.0 \
  --with github.com/phymbert/xk6-sse
mv k6 k6-sse
