# Dockerfile สำหรับโปรเจค go-sse-demo-sensor-dashboard
# ใช้ multistage build เพื่อลดขนาดของ image

# ARG สำหรับรองรับหลาย architecture (linux/amd64, linux/arm64)
ARG GOLANG_VERSION=1.23-alpine
ARG ALPINE_VERSION=latest

# ------------------------------------------------------------------------------
# Stage 1: Build stage - ใช้สำหรับการคอมไพล์โค้ด
# ------------------------------------------------------------------------------
FROM golang:${GOLANG_VERSION} AS builder

# แสดงข้อมูล platform ที่กำลังสร้าง
ARG TARGETPLATFORM
ARG BUILDPLATFORM
RUN echo "Building for $TARGETPLATFORM on $BUILDPLATFORM"

# ติดตั้ง dependencies ที่จำเป็นสำหรับการ build
RUN apk add --no-cache git

# ตั้งค่า working directory
WORKDIR /app

# คัดลอก go.mod และ go.sum เพื่อทำการติดตั้ง dependencies ล่วงหน้า
# (แยกขั้นตอนนี้เพื่อใช้ประโยชน์จาก Docker cache)
COPY go.mod go.sum ./
RUN go mod download

# คัดลอกแต่ละโฟลเดอร์แยกกันเพื่อให้ cache ทำงานได้ดีขึ้น
COPY backend ./backend
COPY frontend ./frontend
COPY configs ./configs

# Build แอปพลิเคชัน สำหรับแต่ละ environment
ARG BUILD_ENV=dev
RUN echo "Building for ${BUILD_ENV} environment" && \
    cd backend && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ../server .

# ------------------------------------------------------------------------------
# Stage 2: UAT image - ใช้สำหรับการ deploy บน UAT
# ------------------------------------------------------------------------------
FROM alpine:${ALPINE_VERSION} AS uat

# ติดตั้ง ca-certificates สำหรับการเชื่อมต่อ HTTPS
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# ตั้งค่า timezone
ENV TZ=Asia/Bangkok
# ตั้งค่า environment
ENV APP_ENV=uat

# ตั้งค่า working directory
WORKDIR /app

# คัดลอกไฟล์ที่จำเป็นจาก builder stage
COPY --from=builder /app/server /app/server
COPY --from=builder /app/frontend /app/frontend
COPY --from=builder /app/configs/backend/.env.uat /app/configs/backend/.env.uat

# ทำให้แน่ใจว่าโฟลเดอร์ต่างๆ มีอยู่
RUN mkdir -p /app/frontend/static /app/configs/backend

# กำหนดสิทธิ์การเรียกใช้งาน
RUN chmod +x /app/server

# เปิด port
EXPOSE 8080

# ตรวจสอบความพร้อมของแอปพลิเคชัน
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

# คำสั่งเริ่มต้นเซิร์ฟเวอร์
CMD ["/app/server"]

# ------------------------------------------------------------------------------
# Stage 3: Production image - ใช้สำหรับการ deploy บน production
# ------------------------------------------------------------------------------
FROM alpine:${ALPINE_VERSION} AS prod

# ติดตั้ง ca-certificates สำหรับการเชื่อมต่อ HTTPS
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# ตั้งค่า timezone
ENV TZ=Asia/Bangkok
# ตั้งค่า environment
ENV APP_ENV=prod

# สร้าง non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# ตั้งค่า working directory
WORKDIR /app

# คัดลอกไฟล์ที่จำเป็นจาก builder stage
COPY --from=builder /app/server /app/server
COPY --from=builder /app/frontend /app/frontend
COPY --from=builder /app/configs/backend/.env.prod /app/configs/backend/.env.prod

# ทำให้แน่ใจว่าโฟลเดอร์ต่างๆ มีอยู่
RUN mkdir -p /app/frontend/static /app/configs/backend

# ตั้งค่า permissions
RUN chown -R appuser:appgroup /app && \
    chmod +x /app/server

# เปลี่ยนเป็น non-root user
USER appuser

# เปิด port
EXPOSE 8080

# ตรวจสอบความพร้อมของแอปพลิเคชัน
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

# คำสั่งเริ่มต้นเซิร์ฟเวอร์
CMD ["/app/server"] 