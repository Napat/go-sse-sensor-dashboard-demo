// สร้างตัวแปรสำหรับเก็บค่าประวัติ
const MAX_DATA_POINTS = 100;

// เก็บข้อมูลประวัติเป็น object ตาม sensor id
const sensorHistory = {};
const timeLabels = [];

// สีสำหรับกราฟใน dark theme
const chartColors = {
    temperature: {
        border: 'rgb(239, 68, 68)',
        background: 'rgba(239, 68, 68, 0.3)'
    },
    humidity: {
        border: 'rgb(56, 189, 248)',
        background: 'rgba(56, 189, 248, 0.3)'
    },
    timestamp: {
        border: 'rgb(251, 191, 36)',
        background: 'rgba(251, 191, 36, 0.3)'
    },
    grid: 'rgba(241, 245, 249, 0.08)',
    text: 'rgba(241, 245, 249, 0.8)'
};

// กำหนดสีสำหรับเซนเซอร์แต่ละตัว (จะถูกกำหนดแบบไดนามิก)
const sensorColorPalette = [
    { border: 'rgb(239, 68, 68)', background: 'rgba(239, 68, 68, 0.3)' },   // แดง
    { border: 'rgb(56, 189, 248)', background: 'rgba(56, 189, 248, 0.3)' }, // ฟ้า
    { border: 'rgb(251, 191, 36)', background: 'rgba(251, 191, 36, 0.3)' }, // เหลือง
    { border: 'rgb(244, 114, 182)', background: 'rgba(244, 114, 182, 0.3)' }, // ชมพู
    { border: 'rgb(34, 197, 94)', background: 'rgba(34, 197, 94, 0.3)' },   // เขียว
    { border: 'rgb(168, 85, 247)', background: 'rgba(168, 85, 247, 0.3)' }, // ม่วง
    { border: 'rgb(234, 179, 8)', background: 'rgba(234, 179, 8, 0.3)' },   // เหลืองอำพัน
    { border: 'rgb(14, 165, 233)', background: 'rgba(14, 165, 233, 0.3)' }  // น้ำเงิน
];

// เตรียมกราฟ
let temperatureChart, humidityChart;

// เก็บข้อมูลเซนเซอร์ล่าสุด
let latestSensors = [];

// ดึงข้อมูล environment จาก API
fetch('/api/environment')
    .then(response => {
        if (!response.ok) {
            throw new Error(`Network response error: ${response.status}`);
        }
        return response.json();
    })
    .then(data => {
        const envBadge = document.getElementById('environment-badge');
        const environment = data.environment;
        
        // แสดงค่า environment
        envBadge.textContent = environment;
        
        // เพิ่ม class ตามสภาพแวดล้อม
        envBadge.classList.add(`env-${environment}`);
        
        // เพิ่ม environment ลงในหัวเพจ
        document.title = `${document.title} (${environment.toUpperCase()})`;
    })
    .catch(error => {
        console.error('Error fetching environment:', error);
        document.getElementById('environment-badge').textContent = 'Unknown';
        showError('ไม่สามารถโหลดข้อมูลสภาพแวดล้อมได้');
    });

// ดึงข้อมูลเซนเซอร์ทั้งหมด
function fetchAllSensors() {
    fetch('/api/sensors')
        .then(response => {
            if (!response.ok) {
                throw new Error(`Network response error: ${response.status}`);
            }
            return response.json();
        })
        .then(sensors => {
            latestSensors = sensors;
            displaySensorsList(sensors);
            updateDashboardCards(sensors);
        })
        .catch(error => {
            console.error('Error fetching all sensors:', error);
            showError('ไม่สามารถโหลดข้อมูลเซนเซอร์ทั้งหมดได้');
        });
}

// อัพเดตข้อมูลบนการ์ดหลัก
function updateDashboardCards(sensors) {
    // หาเซนเซอร์แบบ combined เป็นหลัก
    const combinedSensor = sensors.find(sensor => sensor.type === 'combined');
    
    if (combinedSensor) {
        document.getElementById("temperature").textContent = `${combinedSensor.temperature.toFixed(2)} °C`;
        document.getElementById("humidity").textContent = `${combinedSensor.humidity.toFixed(2)} %`;
        document.getElementById("timestamp").textContent = new Date(combinedSensor.timestamp).toLocaleString();
    } else {
        // ถ้าไม่มี combined sensor ให้หาเซนเซอร์อุณหภูมิและความชื้นแทน
        const tempSensor = sensors.find(sensor => sensor.type === 'temperature');
        const humidSensor = sensors.find(sensor => sensor.type === 'humidity');
        
        if (tempSensor) {
            document.getElementById("temperature").textContent = `${tempSensor.temperature.toFixed(2)} °C`;
            document.getElementById("timestamp").textContent = new Date(tempSensor.timestamp).toLocaleString();
        } else {
            document.getElementById("temperature").textContent = "N/A";
        }
        
        if (humidSensor) {
            document.getElementById("humidity").textContent = `${humidSensor.humidity.toFixed(2)} %`;
            if (!tempSensor) document.getElementById("timestamp").textContent = new Date(humidSensor.timestamp).toLocaleString();
        } else {
            document.getElementById("humidity").textContent = "N/A";
        }
        
        if (!tempSensor && !humidSensor) {
            document.getElementById("timestamp").textContent = "N/A";
        }
    }
}

// แสดงรายการเซนเซอร์ทั้งหมด
function displaySensorsList(sensors) {
    const sensorsListElement = document.getElementById('sensorsList');
    sensorsListElement.innerHTML = '';
    
    if (sensors && sensors.length > 0) {
        sensors.forEach(sensor => {
            const sensorItem = document.createElement('div');
            sensorItem.className = 'sensor-item';
            
            const nameElement = document.createElement('span');
            nameElement.className = 'sensor-item-name';
            nameElement.textContent = `${sensor.name} (${sensor.id})`;
            
            const valuesElement = document.createElement('span');
            
            if (sensor.type === 'temperature') {
                valuesElement.textContent = `${sensor.temperature.toFixed(2)} °C`;
            } else if (sensor.type === 'humidity') {
                valuesElement.textContent = `${sensor.humidity.toFixed(2)} %`;
            } else if (sensor.type === 'combined') {
                valuesElement.textContent = `${sensor.temperature.toFixed(2)} °C / ${sensor.humidity.toFixed(2)} %`;
            }
            
            sensorItem.appendChild(nameElement);
            sensorItem.appendChild(valuesElement);
            sensorsListElement.appendChild(sensorItem);
        });
    } else {
        sensorsListElement.innerHTML = '<p>ไม่พบข้อมูลเซนเซอร์</p>';
    }
}

// สร้างกราฟด้วย Chart.js
function initializeCharts() {
    const tempCtx = document.getElementById('temperatureChart').getContext('2d');
    const humidCtx = document.getElementById('humidityChart').getContext('2d');
    
    // ตั้งค่า global chart options สำหรับ dark theme
    Chart.defaults.color = chartColors.text;
    Chart.defaults.scale.grid.color = chartColors.grid;
    
    // กราฟอุณหภูมิ
    temperatureChart = new Chart(tempCtx, {
        type: 'line',
        data: {
            labels: timeLabels,
            datasets: []
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: false,
                    grid: {
                        color: chartColors.grid
                    },
                    ticks: {
                        color: chartColors.text,
                        font: {
                            weight: 'bold'
                        }
                    },
                    title: {
                        display: true,
                        text: 'Temperature (°C)',
                        color: chartColors.text
                    },
                    // ตั้งค่าช่วงค่าคงที่เพื่อลดการกระโดดของกราฟ
                    suggestedMin: 20,
                    suggestedMax: 35
                },
                x: {
                    grid: {
                        color: chartColors.grid,
                        display: false
                    },
                    ticks: {
                        color: chartColors.text,
                        maxRotation: 45,
                        minRotation: 45,
                        // แสดงเฉพาะบางป้ายกำกับเมื่อมีข้อมูลมาก
                        autoSkip: true,
                        maxTicksLimit: 10
                    }
                }
            },
            plugins: {
                legend: {
                    labels: {
                        color: chartColors.text,
                        font: {
                            weight: 'bold'
                        }
                    },
                    position: 'top'
                },
                tooltip: {
                    mode: 'index',
                    intersect: false
                }
            },
            animation: {
                duration: 300, // ลดเวลา animation ลงเพื่อลดการกระโดด
                easing: 'linear'
            },
            transitions: {
                active: {
                    animation: {
                        duration: 300
                    }
                }
            },
            interaction: {
                mode: 'nearest',
                axis: 'x',
                intersect: false
            },
            elements: {
                line: {
                    tension: 0.3
                },
                point: {
                    radius: 3,
                    hoverRadius: 5
                }
            }
        }
    });
    
    // กราฟความชื้น
    humidityChart = new Chart(humidCtx, {
        type: 'line',
        data: {
            labels: timeLabels,
            datasets: []
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: false,
                    grid: {
                        color: chartColors.grid
                    },
                    ticks: {
                        color: chartColors.text,
                        font: {
                            weight: 'bold'
                        }
                    },
                    title: {
                        display: true,
                        text: 'Humidity (%)',
                        color: chartColors.text
                    },
                    // ตั้งค่าช่วงค่าคงที่เพื่อลดการกระโดดของกราฟ
                    suggestedMin: 40,
                    suggestedMax: 80
                },
                x: {
                    grid: {
                        color: chartColors.grid,
                        display: false
                    },
                    ticks: {
                        color: chartColors.text,
                        maxRotation: 45,
                        minRotation: 45,
                        // แสดงเฉพาะบางป้ายกำกับเมื่อมีข้อมูลมาก
                        autoSkip: true,
                        maxTicksLimit: 10
                    }
                }
            },
            plugins: {
                legend: {
                    labels: {
                        color: chartColors.text,
                        font: {
                            weight: 'bold'
                        }
                    },
                    position: 'top'
                },
                tooltip: {
                    mode: 'index',
                    intersect: false
                }
            },
            animation: {
                duration: 300, // ลดเวลา animation ลงเพื่อลดการกระโดด
                easing: 'linear'
            },
            transitions: {
                active: {
                    animation: {
                        duration: 300
                    }
                }
            },
            interaction: {
                mode: 'nearest',
                axis: 'x',
                intersect: false
            },
            elements: {
                line: {
                    tension: 0.3
                },
                point: {
                    radius: 3,
                    hoverRadius: 5
                }
            }
        }
    });
}

// อัปเดตกราฟด้วยข้อมูลใหม่
function updateCharts(sensors, timestamp) {
    // แปลงวันที่เป็นเวลาท้องถิ่น
    const date = new Date(timestamp);
    const timeLabel = date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    
    // เพิ่มเวลาใหม่
    timeLabels.push(timeLabel);
    
    // จำกัดจำนวนข้อมูลเวลาที่แสดง
    if (timeLabels.length > MAX_DATA_POINTS) {
        timeLabels.shift();
    }
    
    // คำนวณและแสดงช่วงเวลาของข้อมูลที่แสดง
    updateHistoryDuration();
    
    // แบ่งกลุ่มเซนเซอร์ตามประเภท
    const tempSensors = sensors.filter(s => s.type === 'temperature' || s.type === 'combined');
    const humidSensors = sensors.filter(s => s.type === 'humidity' || s.type === 'combined');
    
    // เตรียมข้อมูลสำหรับกราฟอุณหภูมิ
    tempSensors.forEach((sensor, index) => {
        // สร้าง entry ใหม่สำหรับเซนเซอร์ถ้ายังไม่มี
        if (!sensorHistory[sensor.id]) {
            sensorHistory[sensor.id] = {
                data: Array(timeLabels.length - 1).fill(null), // ใส่ค่า null สำหรับประวัติก่อนหน้า
                type: sensor.type
            };
        }
        
        // เพิ่มข้อมูลใหม่
        sensorHistory[sensor.id].data.push(sensor.temperature);
        
        // จำกัดจำนวนข้อมูล
        if (sensorHistory[sensor.id].data.length > MAX_DATA_POINTS) {
            sensorHistory[sensor.id].data.shift();
        }
    });
    
    // เตรียมข้อมูลสำหรับกราฟความชื้น
    humidSensors.forEach((sensor, index) => {
        // สร้าง entry ใหม่สำหรับเซนเซอร์ถ้ายังไม่มี
        if (!sensorHistory[sensor.id]) {
            sensorHistory[sensor.id] = {
                data: Array(timeLabels.length - 1).fill(null), // ใส่ค่า null สำหรับประวัติก่อนหน้า
                type: sensor.type
            };
        }
        
        // เพิ่มข้อมูลใหม่
        sensorHistory[sensor.id].data.push(sensor.humidity);
        
        // จำกัดจำนวนข้อมูล
        if (sensorHistory[sensor.id].data.length > MAX_DATA_POINTS) {
            sensorHistory[sensor.id].data.shift();
        }
    });
    
    // อัปเดตกราฟอุณหภูมิ
    if (temperatureChart) {
        // เตรียม datasets ปัจจุบัน
        const currentDatasets = temperatureChart.data.datasets;
        const currentIds = currentDatasets.map(ds => ds.sensorId);
        
        // ตรวจสอบเซนเซอร์ที่มีอยู่และอัพเดตข้อมูล
        tempSensors.forEach((sensor, index) => {
            const colorIndex = index % sensorColorPalette.length;
            const existingIndex = currentIds.indexOf(sensor.id);
            
            if (existingIndex >= 0) {
                // อัพเดตข้อมูลใน dataset ที่มีอยู่
                currentDatasets[existingIndex].data = sensorHistory[sensor.id].data;
            } else {
                // สร้าง dataset ใหม่สำหรับเซนเซอร์ใหม่
                currentDatasets.push({
                    label: sensor.name,
                    data: sensorHistory[sensor.id].data,
                    borderColor: sensorColorPalette[colorIndex].border,
                    backgroundColor: sensorColorPalette[colorIndex].background,
                    borderWidth: 2,
                    tension: 0.3,
                    fill: 0.1, // ลดความทึบของพื้นที่ใต้กราฟ
                    pointBackgroundColor: sensorColorPalette[colorIndex].border,
                    pointBorderColor: '#fff',
                    pointRadius: 3,
                    pointHoverRadius: 5,
                    sensorId: sensor.id // เก็บ id เพื่อให้ระบุได้ว่าเป็น dataset ของเซนเซอร์ตัวไหน
                });
            }
        });
        
        // ลบ datasets ของเซนเซอร์ที่ไม่มีในข้อมูลใหม่
        temperatureChart.data.datasets = currentDatasets.filter(ds => 
            tempSensors.some(sensor => sensor.id === ds.sensorId)
        );
        
        temperatureChart.data.labels = timeLabels;
        temperatureChart.update('none'); // ใช้ mode 'none' เพื่อข้ามการ animation ในการอัพเดตแบบรวดเร็ว
    }
    
    // อัปเดตกราฟความชื้น
    if (humidityChart) {
        // เตรียม datasets ปัจจุบัน
        const currentDatasets = humidityChart.data.datasets;
        const currentIds = currentDatasets.map(ds => ds.sensorId);
        
        // ตรวจสอบเซนเซอร์ที่มีอยู่และอัพเดตข้อมูล
        humidSensors.forEach((sensor, index) => {
            const colorIndex = index % sensorColorPalette.length;
            const existingIndex = currentIds.indexOf(sensor.id);
            
            if (existingIndex >= 0) {
                // อัพเดตข้อมูลใน dataset ที่มีอยู่
                currentDatasets[existingIndex].data = sensorHistory[sensor.id].data;
            } else {
                // สร้าง dataset ใหม่สำหรับเซนเซอร์ใหม่
                currentDatasets.push({
                    label: sensor.name,
                    data: sensorHistory[sensor.id].data,
                    borderColor: sensorColorPalette[colorIndex].border,
                    backgroundColor: sensorColorPalette[colorIndex].background,
                    borderWidth: 2,
                    tension: 0.3,
                    fill: 0.1, // ลดความทึบของพื้นที่ใต้กราฟ
                    pointBackgroundColor: sensorColorPalette[colorIndex].border,
                    pointBorderColor: '#fff',
                    pointRadius: 3,
                    pointHoverRadius: 5,
                    sensorId: sensor.id // เก็บ id เพื่อให้ระบุได้ว่าเป็น dataset ของเซนเซอร์ตัวไหน
                });
            }
        });
        
        // ลบ datasets ของเซนเซอร์ที่ไม่มีในข้อมูลใหม่
        humidityChart.data.datasets = currentDatasets.filter(ds => 
            humidSensors.some(sensor => sensor.id === ds.sensorId)
        );
        
        humidityChart.data.labels = timeLabels;
        humidityChart.update('none'); // ใช้ mode 'none' เพื่อข้ามการ animation ในการอัพเดตแบบรวดเร็ว
    }
}

// คำนวณระยะเวลาย้อนหลังที่กำลังแสดงข้อมูล
function updateHistoryDuration() {
    if (timeLabels.length <= 1) return;
    
    const tempTitle = document.getElementById('temperature-chart-title');
    const humidTitle = document.getElementById('humidity-chart-title');
    
    if (!tempTitle || !humidTitle) return;
    
    let durationText = '';
    
    if (latestSensors.length > 0) {
        // ใช้จำนวนจุดข้อมูลคูณกับช่วงเวลาโดยประมาณระหว่างข้อมูล (สมมติว่า 5 วินาที)
        // แนะนำให้ปรับค่านี้ตามความถี่จริงของข้อมูลที่ได้รับ
        const dataInterval = 5; // ประมาณช่วงเวลาระหว่างข้อมูล (วินาที)
        const totalSeconds = (timeLabels.length - 1) * dataInterval;
        const pointsCount = timeLabels.length; 
        
        durationText = `แสดงข้อมูลย้อนหลัง `;
        
        if (totalSeconds < 60) {
            durationText += `${totalSeconds} วินาที (${pointsCount} จุด)`;
        } else if (totalSeconds < 3600) {
            const minutes = Math.floor(totalSeconds / 60);
            const seconds = totalSeconds % 60;
            durationText += `${minutes} นาที${seconds > 0 ? ` ${seconds} วินาที` : ''} (${pointsCount} จุด)`;
        } else {
            const hours = Math.floor(totalSeconds / 3600);
            const minutes = Math.floor((totalSeconds % 3600) / 60);
            durationText += `${hours} ชั่วโมง${minutes > 0 ? ` ${minutes} นาที` : ''} (${pointsCount} จุด)`;
        }
    }
    
    tempTitle.textContent = durationText;
    humidTitle.textContent = durationText;
}

// ฟังก์ชั่นแสดงข้อความ error
function showError(message) {
    const errorDiv = document.getElementById('error-message');
    errorDiv.textContent = message;
    errorDiv.style.display = 'block';
    
    // ซ่อนข้อความ error หลังจาก 5 วินาที
    setTimeout(() => {
        errorDiv.style.display = 'none';
    }, 5000);
}

// สร้าง function สำหรับ SSE connection พร้อม retry
function connectSSE() {
    console.log('Connecting to SSE...');
    
    // แสดงสถานะกำลังโหลด
    document.getElementById("temperature").textContent = "กำลังโหลด...";
    document.getElementById("humidity").textContent = "กำลังโหลด...";
    document.getElementById("timestamp").textContent = "กำลังเชื่อมต่อ...";
    document.getElementById("server-id").textContent = "กำลังรอข้อมูล...";
    
    const eventSource = new EventSource('/api/sensors/stream');
    
    eventSource.onopen = function() {
        console.log('SSE connection established');
    };

    eventSource.onmessage = function(event) {
        try {
            // Handle double-encoded JSON string
            let data = JSON.parse(event.data);
            if (typeof data === 'string') {
                data = JSON.parse(data);
            }
            
            // แสดง server ID
            if (data.server_id) {
                document.getElementById("server-id").textContent = data.server_id;
            }
            
            if (data.data && Array.isArray(data.data) && data.data.length > 0) {
                // เก็บข้อมูลล่าสุด
                latestSensors = data.data;
                
                // หาเวลาล่าสุดจากเซนเซอร์
                const timestamp = data.data[0].timestamp;
                
                // อัปเดตค่าที่แสดงบนการ์ด
                updateDashboardCards(data.data);
                
                // อัปเดตกราฟ
                updateCharts(data.data, timestamp);
                
                // อัปเดตรายการเซนเซอร์
                displaySensorsList(data.data);
                
                return;
            }
        } catch (error) {
            console.warn("Error processing sensor data:", error);
        }
        
        // Default values if data is invalid
        document.getElementById("temperature").textContent = "N/A";
        document.getElementById("humidity").textContent = "N/A";
        document.getElementById("timestamp").textContent = "Error loading data";
    };

    eventSource.onerror = function(error) {
        console.error('SSE connection error:', error);
        document.getElementById("server-id").textContent = "ขาดการเชื่อมต่อ";
        // Reconnect will be handled automatically by the browser
    };
    
    return eventSource;
}

// สร้างกราฟเมื่อโหลดหน้าเว็บ
document.addEventListener('DOMContentLoaded', function() {
    initializeCharts();
    fetchAllSensors();
});

// เริ่มการเชื่อมต่อ SSE
const eventSource = connectSSE();

// เพิ่ม event listener สำหรับเมื่อมีการปิดหน้าเว็บ
window.addEventListener('beforeunload', () => {
    if (eventSource) {
        eventSource.close();
    }
});