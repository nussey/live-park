#include <RH_ASK.h>
#include <SPI.h>
#define N_POINTS 100
#define HEARTBEAT_PERIOD  30000

RH_ASK driver;
int sensorpin = 0;
int val;
float oldvoltage[N_POINTS];
bool state = 1;
bool previous_state = 1;
bool is_changed = 0;
unsigned long timer;
float batteryPercentage = 100.0;


void setup() {
  Serial.begin(9600);
  Serial.println("FIRST I PARK MY CAR.........");
  if (!driver.init()) {
    Serial.println("init failed");
  }
  
  for(int i = 0; i < N_POINTS; i++) {
    val = analogRead(sensorpin);
    oldvoltage[i] = 0;
  }
}

struct _packet {
  uint32_t identifier;
  uint8_t occupied;
  uint8_t battery_percentage;
} packetdata;

float voltage;
long lastUpdate = 0;
void loop() {
  timer = millis();
  
  if (timer%10 == 0) {
    val = analogRead(sensorpin);
    batteryPercentage -= .01 * .01;
    if(batteryPercentage < 0)
      batteryPercentage = 0;

    voltage = 5.0 / 1024 * val;
    uint8_t j;
    for(j = 1; j < N_POINTS; j++) {
      oldvoltage[j-1] = oldvoltage[j];
    }
    oldvoltage[N_POINTS-1] = voltage;
    for(j = 0; j < N_POINTS; j++) {
      voltage += oldvoltage[j];
    }

    voltage /= (N_POINTS + 1);
    if(state) {
      if(voltage <= 1.0 && oldvoltage[N_POINTS/2] > voltage) {
        state = 0;
      }
    } else {
      if(voltage >= 1.0 && oldvoltage[N_POINTS/2] < voltage) {
        state = 1;
      }
    }
    
    packetdata.identifier = 0x69696969;
    packetdata.occupied = state;
    packetdata.battery_percentage = batteryPercentage / 100.0 * 255;
  
    is_changed = !(previous_state == state);
  
    Serial.println(voltage);
    
    if (is_changed) {
        driver.send((uint8_t*)&packetdata,sizeof(struct _packet));
        //Serial.println("SENT");
        if (state) {
          Serial.println("PARKED");
        } else {
          Serial.println("EMPTY");
        }
        Serial.println(voltage);
        previous_state = state;
    } 
    if (timer % HEARTBEAT_PERIOD == 0) {
        driver.send((uint8_t*)&packetdata,sizeof(struct _packet));
        Serial.println("SENT EVERY 30");
        delay(1);
        //Serial.println(timer/1000);
    }
  }
}
