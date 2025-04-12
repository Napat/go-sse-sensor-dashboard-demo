.PHONY: all
all: curl-dev

# ค่าเริ่มต้นสำหรับ development environment
HOST_DEV ?= localhost
PORT_DEV ?= 8081
VUS_DEV ?= 2
DURATION_DEV ?= 10s

# ค่าเริ่มต้นสำหรับ UAT environment
HOST_UAT ?= localhost
PORT_UAT ?= 8082
VUS_UAT ?= 700
DURATION_UAT ?= 20s

# ค่าเริ่มต้นสำหรับ production environment
HOST_PROD ?= localhost
PORT_PROD ?= 8083
VUS_PROD ?= 1700
DURATION_PROD ?= 10m

.PHONY: curl-dev
curl-dev:
	curl -N -H "Accept: text/event-stream" http://$(HOST_DEV):$(PORT_DEV)/api/sensors/stream

.PHONY: curl-uat
curl-uat:
	curl -N -H "Accept: text/event-stream" http://$(HOST_UAT):$(PORT_UAT)/api/sensors/stream

.PHONY: curl-prod
curl-prod:
	curl -N -H "Accept: text/event-stream" http://$(HOST_PROD):$(PORT_PROD)/api/sensors/stream
	
.PHONY: pre-loadtest-dev
pre-loadtest-dev:
	./k6-sse run -e HOST=$(HOST_DEV) -e PORT=$(PORT_DEV) --vus 2 --duration 10s ./sse-loadtest.js
	
.PHONY: pre-loadtest-uat
pre-loadtest-uat:
	./k6-sse run -e HOST=$(HOST_UAT) -e PORT=$(PORT_UAT) --vus 2 --duration 10s ./sse-loadtest.js

.PHONY: pre-loadtest-prod
pre-loadtest-prod:
	./k6-sse run -e HOST=$(HOST_PROD) -e PORT=$(PORT_PROD) --vus 2 --duration 10s ./sse-loadtest.js
	
.PHONY: loadtest-dev
loadtest-dev:
	./k6-sse run -e HOST=$(HOST_DEV) -e PORT=$(PORT_DEV) --vus $(VUS_DEV) --duration $(DURATION_DEV) ./sse-loadtest.js

.PHONY: loadtest-dev-dashboard
loadtest-dev-dashboard:
	./k6-sse run --out dashboard -e HOST=$(HOST_DEV) -e PORT=$(PORT_DEV) --vus $(VUS_DEV) --duration $(DURATION_DEV) ./sse-loadtest.js

.PHONY: loadtest-uat
loadtest-uat:
	./k6-sse run -e HOST=$(HOST_UAT) -e PORT=$(PORT_UAT) --vus $(VUS_UAT) --duration $(DURATION_UAT) ./sse-loadtest.js

.PHONY: loadtest-uat-dashboard
loadtest-uat-dashboard:
	./k6-sse run --out 'dashboard=port=5666&open=true' -e HOST=$(HOST_UAT) -e PORT=$(PORT_UAT) --vus $(VUS_UAT) --duration $(DURATION_UAT) ./sse-loadtest.js

.PHONY: loadtest-prod
loadtest-prod:
	./k6-sse run -e HOST=$(HOST_PROD) -e PORT=$(PORT_PROD) --vus $(VUS_PROD) --duration $(DURATION_PROD) ./sse-loadtest.js

.PHONY: loadtest-prod-dashboard
loadtest-prod-dashboard:
	./k6-sse run --out 'dashboard=port=5667&open=true' -e HOST=$(HOST_PROD) -e PORT=$(PORT_PROD) --vus $(VUS_PROD) --duration $(DURATION_PROD) ./sse-loadtest.js

# ทำ loadtest โดยใช้ custom parameters
# ตัวอย่างการใช้: make loadtest-custom HOST=my-api.com PORT=9000 VUS=50 DURATION=5m
.PHONY: loadtest-custom
loadtest-custom:
	./k6-sse run -e HOST=$(HOST) -e PORT=$(PORT) --vus $(VUS) --duration $(DURATION) ./sse-loadtest.js

# ติดตั้งเครื่องมือสำหรับ development
.PHONY: install-dev-tools
install-dev-tools:
	./k6sse_build_mac.sh
