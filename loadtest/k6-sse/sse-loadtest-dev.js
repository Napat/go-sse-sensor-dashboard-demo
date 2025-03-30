import sse from "k6/x/sse"
import {check} from "k6"
import {__ENV} from "k6/execution"

const timestamp = () => `[${new Date().toUTCString()}]`
const log = (...args) => console.log(timestamp(), ...args)

// ฟังก์ชันสำหรับตรวจสอบ timeout และปิดการเชื่อมต่อจากฝั่ง client
function checkTimeoutAndClose(client, startTime, maxDuration, shouldCloseRef) {
  const currentTime = new Date().getTime();
  const elapsedTime = currentTime - startTime;
  
  // ถ้าเวลาผ่านไปเกิน maxDuration ให้ปิดการเชื่อมต่อ
  if (elapsedTime >= maxDuration && !shouldCloseRef.value) {
      log(`elapsed time: ${elapsedTime}ms`);
      log('closing connection due to timeout');
      shouldCloseRef.value = true;
      client.close();
      return true;
  }
  return false;
}

export default function () {
    // รับค่า HOST และ PORT จาก environment variables หรือใช้ค่าเริ่มต้น
    const host = __ENV.HOST || 'localhost';
    const port = __ENV.PORT || '8081';
    const url = `http://${host}:${port}/api/sensors/stream`;
    
    log(`Connecting to ${url}`);
    
    const params = {
        method: 'GET',
        headers: {
          'Accept': 'text/event-stream',
          'Cache-Control': 'no-cache',
        }
    }

    const startTime = new Date().getTime();
    const maxDuration = 5000;
    let shouldCloseRef = { value: false };  // ใช้ object เพื่อให้สามารถส่งค่า reference ได้

    const response = sse.open(url, params, function (client) {
        client.on('open', function open() {
          log(`connected`)
        })

        client.on('event', function (event) {
          log(`event id=${event.id}, name=${event.name}, data=${event.data}`)

            // k6-sse limitation
            // The module will not support async io and the javascript main loop will be blocked during the http request duration.
            // ref: https://github.com/phymbert/xk6-sse/blob/main/docs/design/021-sse-api.md?plain=1#L26
            // เราจึงต้องมาเช็ค timeout ได้เฉพาะใน event callback ที่ k6-sse เรียกเข้ามาเท่านั้น ถ้าไปทำที่อื่น code จะไม่ถูกเรียกทำงาน
            checkTimeoutAndClose(client, startTime, maxDuration, shouldCloseRef);
        })

        client.on('error', function (e) {
          log('An unexpected error occurred: ', e.error())
        })
    })

    check(response, {"status is 200": (r) => r && r.status === 200})
}
