# Server-Sent Events (SSE)

## SSE คืออะไร

Server-Sent Events (SSE) เป็นเทคโนโลยีที่ช่วยให้เซิร์ฟเวอร์สามารถส่งข้อมูลแบบเรียลไทม์มายังเว็บเบราว์เซอร์ของผู้ใช้งานผ่านการเชื่อมต่อ HTTP ที่เปิดไว้เพียงครั้งเดียว โดยไม่จำเป็นต้องให้ไคลเอนต์ส่งคำขอใหม่ซ้ำๆ SSE ถูกออกแบบมาเพื่อการสื่อสารทิศทางเดียว (one-way communication) จากเซิร์ฟเวอร์ไปยังไคลเอนต์ ทำให้เหมาะสำหรับการส่งการอัปเดตแบบ real-time เช่น การแจ้งเตือน, ข้อมูลตลาดหุ้น, อัพเดทสถานะ หรือ feed ข่าวสด

ด้วยการทำงานบนพื้นฐานของโปรโตคอล HTTP ทั่วไป ทำให้ SSE มีความเรียบง่ายในการใช้งานและการติดตั้ง โดยไม่จำเป็นต้องใช้โปรโตคอลพิเศษหรือการตั้งค่าเซิร์ฟเวอร์ที่ซับซ้อน

## รูปแบบการทำงาน

การทำงานของ SSE มีลักษณะดังนี้:

1. ไคลเอนต์เริ่มต้นการเชื่อมต่อโดยส่ง HTTP request ไปยัง URL ที่ให้บริการ SSE
2. เซิร์ฟเวอร์ตอบกลับด้วย content type พิเศษคือ `text/event-stream` และเปิดการเชื่อมต่อไว้
3. เมื่อมีข้อมูลใหม่ เซิร์ฟเวอร์จะส่งข้อมูลผ่านการเชื่อมต่อที่เปิดไว้แล้วโดยไม่ปิดการเชื่อมต่อ
4. เซิร์ฟเวอร์สามารถส่งข้อมูลแบบต่อเนื่องเป็นระยะๆ ตามที่ต้องการ
5. การเชื่อมต่อจะยังคงเปิดไว้จนกว่าจะถูกปิดโดยไคลเอนต์หรือเซิร์ฟเวอร์

ข้อมูลที่ส่งผ่าน SSE มีรูปแบบที่เรียกว่า "event stream" ซึ่งเป็นข้อความ UTF-8 ที่มีโครงสร้างเฉพาะ:

```txt
data: ข้อความที่ต้องการส่ง

event: event_name
data: ข้อมูลเกี่ยวกับ event นี้

id: 12345
data: ข้อมูลที่มี id กำกับ
```

## เปรียบเทียบกับเทคโนโลยีอื่น

### SSE vs WebSocket

| คุณลักษณะ | SSE | WebSocket |
|----------|-----|-----------|
| ทิศทางการสื่อสาร | ทิศทางเดียว (เซิร์ฟเวอร์ไปไคลเอนต์) | สองทิศทาง (full-duplex) |
| โปรโตคอล | HTTP/HTTPS มาตรฐาน | โปรโตคอลเฉพาะ (ws/wss) |
| ความซับซ้อนในการติดตั้ง | ต่ำ (ใช้เซิร์ฟเวอร์ HTTP ปกติได้) | สูงกว่า (ต้องการการตั้งค่าเพิ่มเติม) |
| รูปแบบข้อมูล | ข้อความ UTF-8 เท่านั้น | สนับสนุนทั้งข้อความและไบนารี |
| การเชื่อมต่อใหม่อัตโนมัติ | มี (built-in) | ต้องพัฒนาเพิ่มเติม |
| การขยายระบบ | ง่ายกว่า เนื่องจากใช้ HTTP ปกติ | ซับซ้อนกว่า ต้องจัดการการเชื่อมต่อหลายๆ อัน |
| ประสิทธิภาพ | ดีสำหรับการส่งข้อมูลทิศทางเดียว | ดีกว่าสำหรับการสื่อสารแบบโต้ตอบ |

### SSE vs Long Polling

| คุณลักษณะ | SSE | Long Polling |
|----------|-----|--------------|
| การเชื่อมต่อ | เชื่อมต่อเดียวที่เปิดค้างไว้ | ต้องเปิดการเชื่อมต่อใหม่ทุกครั้งที่ได้รับข้อมูล |
| ประสิทธิภาพ | สูงกว่า (overhead ต่ำกว่า) | ต่ำกว่า (overhead สูงจากการเปิดเชื่อมต่อใหม่) |
| การใช้ทรัพยากร | ใช้น้อยกว่า | ใช้มากกว่าจากการสร้างการเชื่อมต่อซ้ำๆ |
| การจัดการข้อมูล | มีการจัดการขอบเขตข้อความและ ID | ต้องพัฒนาเพิ่มเติม |
| ความเข้ากันได้ | ไม่รองรับ IE | รองรับทุกเบราว์เซอร์ |

## ข้อดีของ SSE

1. **เรียบง่าย**: ง่ายต่อการพัฒนาและติดตั้ง เนื่องจากใช้ HTTP ทั่วไป
2. **การเชื่อมต่อใหม่อัตโนมัติ**: เบราว์เซอร์จัดการการเชื่อมต่อใหม่เองเมื่อการเชื่อมต่อถูกตัด
3. **เข้ากันได้กับโครงสร้างเน็ตเวิร์คที่มีอยู่**: ผ่าน proxy และไฟร์วอลล์ได้ง่ายกว่า WebSocket
4. **การจัดการข้อมูล Event**: รองรับ event ID และการกลับมาเชื่อมต่อใหม่แบบต่อเนื่อง
5. **ใช้ทรัพยากรน้อย**: ใช้ทรัพยากรน้อยกว่า WebSocket สำหรับการสื่อสารทิศทางเดียว

## ข้อจำกัดของ SSE

1. **สื่อสารทิศทางเดียว**: ไม่รองรับการส่งข้อมูลจากไคลเอนต์ไปเซิร์ฟเวอร์ (ต้องใช้ XMLHttpRequest หรือ fetch เพิ่มเติม)
2. **จำกัดเฉพาะข้อความ**: ไม่สามารถส่งข้อมูลไบนารีได้โดยตรง (ต้องแปลงเป็น Base64 หากจำเป็น)
3. **จำนวนการเชื่อมต่อจำกัด**: เบราว์เซอร์มักจำกัดจำนวนการเชื่อมต่อ SSE ต่อ domain
4. **ไม่รองรับบางเบราว์เซอร์**: IE ไม่รองรับ SSE โดยตรง (ต้องใช้ polyfill)
5. **การบีบอัดข้อมูล**: การบีบอัดแบบ gzip อาจมีปัญหาในบางสถานการณ์

## เมื่อไรควรใช้ SSE

SSE เหมาะกับกรณีการใช้งานดังต่อไปนี้:

1. **การแจ้งเตือนแบบเรียลไทม์**: การแจ้งเตือนผู้ใช้งานเกี่ยวกับเหตุการณ์ใหม่ๆ หรือการอัปเดต
2. **แดชบอร์ดแสดงข้อมูลสด**: แสดงข้อมูลทางการเงิน สถิติ หรือข้อมูลอื่นๆ ที่อัปเดตเป็นประจำ
3. **อัปเดตโซเชียล**: ฟีดข่าว สถานะ หรือความคิดเห็นที่อัปเดตอย่างต่อเนื่อง
4. **การรายงานความคืบหน้า**: แสดงสถานะของกระบวนการที่ทำงานอยู่บนเซิร์ฟเวอร์
5. **การส่งข้อมูลแคช**: แจ้งไคลเอนต์เมื่อข้อมูลในแคชไม่ถูกต้องและควรรีเฟรช

ควรพิจารณาใช้ WebSocket แทน SSE เมื่อ

1. ต้องการการสื่อสารสองทิศทางแบบเรียลไทม์
2. ต้องการส่งข้อมูลไบนารีขนาดใหญ่
3. ต้องการลดความล่าช้าให้น้อยที่สุดสำหรับแอปพลิเคชันเกมหรือการสื่อสารแบบโต้ตอบสูง

## ตัวอย่างการใช้งาน

### ตัวอย่างฝั่งไคลเอนต์ (JavaScript)

```javascript
// สร้างการเชื่อมต่อ SSE
const eventSource = new EventSource('/events');

// รับข้อมูลทั่วไป
eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('ได้รับข้อมูล:', data);
  updateUI(data);
};

// รับข้อมูลเฉพาะประเภท
eventSource.addEventListener('sensor-update', (event) => {
  const sensorData = JSON.parse(event.data);
  console.log('ข้อมูลเซ็นเซอร์อัปเดต:', sensorData);
  updateSensorDisplay(sensorData);
});

// จัดการเหตุการณ์การเชื่อมต่อ
eventSource.onopen = () => {
  console.log('การเชื่อมต่อ SSE เปิดแล้ว');
};

eventSource.onerror = (error) => {
  console.error('เกิดข้อผิดพลาดในการเชื่อมต่อ SSE:', error);
  // อาจจะลองเชื่อมต่อใหม่หรือแจ้งผู้ใช้
};

// ปิดการเชื่อมต่อเมื่อไม่ใช้แล้ว
function closeConnection() {
  eventSource.close();
}
```

### ตัวอย่างฝั่งเซิร์ฟเวอร์ (Go)

```go
func sseHandler(w http.ResponseWriter, r *http.Request) {
    // ตั้งค่าส่วนหัวสำหรับ SSE
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")  // ในการใช้งานจริงควรระบุเฉพาะโดเมนที่อนุญาตเท่านั้น

    // ส่ง ID เริ่มต้น
    fmt.Fprintf(w, "id: %d\n", time.Now().Unix())
    
    // สร้าง ticker สำหรับส่งข้อมูลทุก 2 วินาที
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    ctx := r.Context() // Use request context
    
    for {
        select {
        case <-ctx.Done(): 
            log.Println("Client disconnected")
            return
        case t := <-ticker.C:
            // สร้างข้อมูลเซ็นเซอร์จำลอง
            sensorData := map[string]interface{}{
                "temperature": rand.Float64()*30.0 + 10.0,
                "humidity": rand.Float64()*60.0 + 30.0,
                "timestamp": t.Unix(),
            }
            
            // แปลงเป็น JSON
            jsonData, _ := json.Marshal(sensorData)
            
            // ส่งข้อมูลพร้อม event type
            fmt.Fprintf(w, "event: sensor-update\n")
            fmt.Fprintf(w, "data: %s\n\n", jsonData)
            
            // Flush ข้อมูลเพื่อให้แน่ใจว่าถูกส่งไปทันที
            if f, ok := w.(http.Flusher); ok {
                f.Flush()
            }
        }
    }
}
```

## การใช้งาน Event ID ใน SSE

Event ID เป็นคุณสมบัติสำคัญของ SSE ที่ช่วยในการจัดการข้อมูลและการเชื่อมต่อใหม่ (reconnection) โดยมีประโยชน์ดังนี้:

### หลักการทำงานของ Event ID

1. **การติดตามลำดับข้อความ**: Event ID ช่วยให้ทั้งเซิร์ฟเวอร์และไคลเอนต์สามารถติดตามข้อความที่ส่งล่าสุดได้
2. **การเชื่อมต่อใหม่อัตโนมัติ**: เมื่อการเชื่อมต่อถูกตัด เบราว์เซอร์จะส่ง ID ของข้อความล่าสุดที่ได้รับในส่วนหัว `Last-Event-ID` เมื่อทำการเชื่อมต่อใหม่
3. **การส่งข้อความที่หายไป**: เซิร์ฟเวอร์สามารถใช้ ID นี้เพื่อส่งเฉพาะข้อความที่ไคลเอนต์ยังไม่ได้รับหลังจากเชื่อมต่อใหม่

### ตัวอย่างการใช้งาน Event ID

#### ตัวอย่างฝั่งไคลเอนต์ (JavaScript) - การใช้งาน Event ID

```javascript
// สร้างการเชื่อมต่อ SSE
const eventSource = new EventSource('/events');

// เก็บ ID ล่าสุดที่ได้รับ (จะถูกเก็บโดยเบราว์เซอร์อัตโนมัติ)
let lastReceivedId = '';

// รับข้อมูลจาก event stream
eventSource.onmessage = (event) => {
  // event.lastEventId จะมีค่า ID ล่าสุดที่ได้รับจากเซิร์ฟเวอร์
  lastReceivedId = event.lastEventId;
  
  console.log(`ได้รับข้อความ ID: ${event.lastEventId}, ข้อมูล: ${event.data}`);
  
  // อัปเดต UI ด้วยข้อมูลใหม่
  updateUI(event.data, event.lastEventId);
};

// เมื่อเกิดข้อผิดพลาดและเชื่อมต่อใหม่ เบราว์เซอร์จะส่ง Last-Event-ID ให้เซิร์ฟเวอร์อัตโนมัติ
eventSource.onerror = (error) => {
  console.error('เกิดข้อผิดพลาดในการเชื่อมต่อ SSE:', error);
  console.log(`กำลังพยายามเชื่อมต่อใหม่... Last Event ID: ${lastReceivedId}`);
};

// แสดงข้อมูลในหน้าเว็บ
function updateUI(data, id) {
  const messageContainer = document.getElementById('messages');
  const messageElement = document.createElement('div');
  messageElement.textContent = `[ID: ${id}] ${data}`;
  messageContainer.appendChild(messageElement);
}
```

#### ตัวอย่างฝั่งเซิร์ฟเวอร์ (Go) - การจัดการ Event ID

```go
func sseHandler(w http.ResponseWriter, r *http.Request) {
    // ตั้งค่าส่วนหัวสำหรับ SSE
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    
    // อ่าน Last-Event-ID จาก request header (ถ้ามี)
    lastEventID := r.Header.Get("Last-Event-ID")
    var lastID int
    
    if lastEventID != "" {
        // แปลง string เป็น int
        var err error
        lastID, err = strconv.Atoi(lastEventID)
        if err != nil {
            lastID = 0
        }
    }
    
    // สมมติว่ามีฟังก์ชัน getMessagesSince(id) ที่ดึงข้อความตั้งแต่ ID ที่ระบุ
    messages := getMessagesSince(lastID)
    
    // ส่งข้อความที่ยังไม่ได้รับก่อน
    for _, msg := range messages {
        fmt.Fprintf(w, "id: %d\n", msg.ID)
        fmt.Fprintf(w, "data: %s\n\n", msg.Data)
        
        if f, ok := w.(http.Flusher); ok {
            f.Flush()
        }
    }
    
    // ตั้ง event ID เริ่มต้นสำหรับข้อความใหม่
    currentID := time.Now().Unix()
    if len(messages) > 0 {
        currentID = messages[len(messages)-1].ID + 1
    }
    
    // สร้าง ticker สำหรับส่งข้อมูลทุก 2 วินาที
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()
    
    ctx := r.Context()
    
    for {
        select {
        case <-ctx.Done():
            log.Println("Client disconnected")
            return
        case <-ticker.C:
            // สร้างข้อมูลใหม่
            data := fmt.Sprintf("ข้อมูลใหม่เวลา %s", time.Now().Format("15:04:05"))
            
            // ส่งข้อมูลพร้อม ID
            fmt.Fprintf(w, "id: %d\n", currentID)
            fmt.Fprintf(w, "data: %s\n\n", data)
            
            // เพิ่ม ID สำหรับข้อความถัดไป
            currentID++
            
            if f, ok := w.(http.Flusher); ok {
                f.Flush()
            }
        }
    }
}

// โครงสร้างข้อมูลจำลองสำหรับตัวอย่าง
type Message struct {
    ID   int
    Data string
}

// ฟังก์ชันจำลองสำหรับดึงข้อความที่ยังไม่ได้ส่ง
func getMessagesSince(id int) []Message {
    // ในระบบจริง นี่อาจจะเป็นการดึงข้อมูลจากฐานข้อมูลหรือแคช
    messages := []Message{
        {1, "ข้อความที่ 1"},
        {2, "ข้อความที่ 2"},
        {3, "ข้อความที่ 3"},
    }
    
    result := []Message{}
    for _, msg := range messages {
        if msg.ID > id {
            result = append(result, msg)
        }
    }
    
    return result
}
```

### ประโยชน์ของการใช้ Event ID

1. **ความต่อเนื่องของข้อมูล**: ช่วยให้ไคลเอนต์ไม่พลาดข้อมูลแม้เชื่อมต่อหลุดชั่วคราว
2. **ประสิทธิภาพ**: ลดปริมาณข้อมูลที่ต้องส่งซ้ำเมื่อเชื่อมต่อใหม่
3. **การติดตาม**: สามารถใช้ ID เป็นตัวอ้างอิงสำหรับการเรียงลำดับหรือตรวจสอบความสมบูรณ์ของข้อมูล
4. **สถานะ**: เก็บสถานะล่าสุดของไคลเอนต์โดยไม่ต้องใช้ session หรือ cookie เพิ่มเติม

## ข้อจำกัดของจำนวนการเชื่อมต่อ SSE

การใช้งาน SSE มีข้อจำกัดเกี่ยวกับจำนวนการเชื่อมต่อที่สามารถทำได้พร้อมกัน ทั้งทางฝั่งเบราว์เซอร์และเซิร์ฟเวอร์:

### ข้อจำกัดฝั่งเบราว์เซอร์

เบราว์เซอร์ส่วนใหญ่มีการจำกัดจำนวนการเชื่อมต่อ HTTP แบบพร้อมกันต่อโดเมน โดยทั่วไปเป็นดังนี้:

| เบราว์เซอร์ | การเชื่อมต่อสูงสุดต่อโดเมน* |
|------------|---------------------------|
| Chrome     | 6                         |
| Firefox    | 6                         |
| Safari     | 6                         |
| Edge       | 6                         |

```plaintext
ref: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#listening_for_custom_events
Warning: When not used over HTTP/2, SSE suffers from a limitation to the maximum number of open connections, which can be especially painful when opening multiple tabs, as the limit is per browser and is set to a very low number (6). The issue has been marked as "Won't fix" in Chrome and Firefox. This limit is per browser + domain, which means that you can open 6 SSE connections across all of the tabs to www.example1.com and another 6 SSE connections to www.example2.com (per Stack Overflow). When using HTTP/2, the maximum number of simultaneous HTTP streams is negotiated between the server and the client (defaults to 100).
```

นั่นหมายความว่าถ้ามีการเชื่อมต่อ SSE เปิดอยู่ 6 การเชื่อมต่อไปยังโดเมนเดียวกัน การร้องขอ HTTP อื่นๆ จะต้องรอจนกว่าจะมีการเชื่อมต่อว่าง ข้อจำกัดนี้รวมถึงการเชื่อมต่อ SSE, XHR, fetch, และการโหลดทรัพยากรอื่นๆ

#### วิธีการตรวจสอบและแก้ไขข้อจำกัดในเบราว์เซอร์

ตรวจสอบจำนวนการเชื่อมต่อปัจจุบัน  

* ใน Chrome: เปิด DevTools > Network tab > แสดงการเชื่อมต่อที่กำลังใช้งาน
* ใน Firefox: เปิด about:networking#sockets ในแถบ URL

แนวทางแก้ไข  

* ใช้เทคนิค Domain Sharding: แยกการเชื่อมต่อไปยังหลาย subdomain (api1.example.com, api2.example.com)
* รวมข้อมูลจากหลาย stream ไว้ในการเชื่อมต่อเดียว และแยกประเภทด้วย event type
* ใช้การเชื่อมต่อ SSE แบบรวมศูนย์และกระจายข้อมูลผ่าน localStorage หรือ SharedWorker ไปยังแท็บอื่นๆ

```javascript
// ตัวอย่างการใช้ SharedWorker เพื่อแชร์การเชื่อมต่อ SSE
// ในไฟล์ sse-worker.js
self.addEventListener('connect', function(e) {
  const port = e.ports[0];
  
  // สร้างการเชื่อมต่อ SSE ครั้งเดียวในเวอร์คเกอร์
  if (!self.eventSource) {
    self.eventSource = new EventSource('/events');
    self.eventSource.onmessage = function(event) {
      // ส่งข้อมูลไปยังทุกพอร์ตที่เชื่อมต่อ (ทุกแท็บ/หน้า)
      self.clients.forEach(client => {
        client.postMessage({
          type: 'message',
          data: event.data,
          lastEventId: event.lastEventId
        });
      });
    };
    
    self.clients = new Set();
  }
  
  self.clients.add(port);
  
  port.addEventListener('message', function(e) {
    // รับคำสั่งจากหน้าเว็บ
    if (e.data.command === 'close') {
      self.clients.delete(port);
      
      // ถ้าไม่มีไคลเอนต์เหลือ ปิดการเชื่อมต่อ
      if (self.clients.size === 0 && self.eventSource) {
        self.eventSource.close();
        self.eventSource = null;
      }
    }
  });
  
  port.start();
});
```

```javascript
// ใช้งาน SharedWorker ในหน้าเว็บ
const worker = new SharedWorker('sse-worker.js');

worker.port.addEventListener('message', function(e) {
  if (e.data.type === 'message') {
    console.log('ได้รับข้อมูล:', e.data.data);
    console.log('Event ID:', e.data.lastEventId);
    // อัปเดต UI
  }
});

worker.port.start();

// เมื่อปิดหน้า
window.addEventListener('beforeunload', function() {
  worker.port.postMessage({command: 'close'});
});
```

## สรุป

Server-Sent Events (SSE) เป็นเทคโนโลยีที่มีประสิทธิภาพสำหรับการส่งข้อมูลแบบเรียลไทม์จากเซิร์ฟเวอร์ไปยังไคลเอนต์ ด้วยความเรียบง่ายในการติดตั้งและการใช้งาน ทำให้เป็นทางเลือกที่ดีสำหรับการพัฒนาแอปพลิเคชันที่ต้องการอัปเดตข้อมูลแบบทิศทางเดียวอย่างต่อเนื่อง

แม้ว่า WebSocket จะมีความสามารถมากกว่าในการสื่อสารสองทิศทาง แต่ SSE ก็มีข้อได้เปรียบในด้านความเรียบง่าย ความเข้ากันได้กับโครงสร้างพื้นฐานที่มีอยู่ และการใช้ทรัพยากรที่น้อยกว่า ทำให้เป็นทางเลือกที่เหมาะสมสำหรับหลายกรณีการใช้งาน

การเลือกระหว่าง SSE และเทคโนโลยีอื่นๆ ควรพิจารณาจากความต้องการเฉพาะของแอปพลิเคชัน โดยคำนึงถึงรูปแบบการสื่อสาร ปริมาณข้อมูล ความถี่ในการอัปเดต และข้อจำกัดด้านทรัพยากร
