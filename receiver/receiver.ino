#include "./libs/VirtualWire/VirtualWire.cpp"
#include "./libs/ArduinoJson/ArduinoJson.h"

#define PACKET_LENGTH     11      // 11 bytes after preamble

#define STATE_IDLE        0
#define STATE_PREAMBLE_1  1
#define STATE_PREAMBLE_2  2
#define STATE_PREAMBLE_3  3
#define STATE_RECV        4


uint8_t buf[16];
uint8_t bufLen = 16;

// State machine variables
uint8_t state = STATE_IDLE;
uint8_t bytesReceived = 0;

struct _packet {
  uint32_t identifier;
  uint8_t occupied;
  uint8_t batteryPercentage;
} packetData;
uint8_t* packetPtr = (uint8_t*)&packetData;

void setup() {
    // start serial port at 9600 bps:
    Serial.begin(9600);
    while (!Serial) {
        ; // wait for serial port to connect. Needed for native USB port only
    }
    
    vw_set_ptt_inverted(true);
    vw_setup(2000);
    vw_rx_start();
}

void reverse(uint8_t *src, uint8_t len) {
  uint8_t tmp;
  for(uint8_t j = 0; j < len; j++) {
    tmp = src[j];
    src[j] = src[len-j-1];
    src[len-j-1] = tmp;
  }
}

char stringBuf[100];
void loop() {
  uint8_t i;
  if(vw_get_message(buf, &bufLen)) {
    /*Serial.println("Got message: ");
     for (i = 0; i < bufLen; i++) {
      Serial.print(buf[i], HEX);
      Serial.print(" ");
     }
     Serial.println();*/
    packetData.identifier = *((uint32_t*)(buf+4));
    packetData.occupied = buf[8];
    packetData.batteryPercentage = buf[9];
    
    StaticJsonBuffer<101> jsonBuffer;
    JsonObject& obj = jsonBuffer.createObject();
    obj["identifier"] = packetData.identifier;
    obj["batteryPercentage"] = packetData.batteryPercentage;
    obj["occupied"] = packetData.occupied;

    obj.printTo(Serial);
    Serial.println();
  }

  /*
  for(i = 0; i < bufLen; i++) {
    switch(state) {
      case STATE_IDLE:
        if(buf[i] == 0x50) {
          state = STATE_PREAMBLE_1;
        }
        break;
      case STATE_PREAMBLE_1:
        if(buf[i] == 0x41) {
          state = STATE_PREAMBLE_2;
        } else {
          state = STATE_IDLE;
        }
        break;
      case STATE_PREAMBLE_2:
        if(buf[i] == 0x52) {
          state = STATE_PREAMBLE_3;
        } else {
          state = STATE_IDLE;
        }
      case STATE_PREAMBLE_3:
        if(buf[i] == 0x4B) {
          state = STATE_RECV;
        } else {
          state = STATE_IDLE;
        }
        break;
      case STATE_RECV:
        if(bytesReceived < PACKET_LENGTH) {

          // Got a full packet!
          packetPtr[bytesReceived] = buf[i];
          bytesReceived++;

          // Check CRC here
        } else {
          state = STATE_IDLE;
          bytesReceived = 0;
        }
    }
    */
}
