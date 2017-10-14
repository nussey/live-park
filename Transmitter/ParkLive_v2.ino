#include <RH_ASK.h>
#include <SPI.h>

RH_ASK driver;
#define n_points 50
int sensorpin = 0;
int val;
float oldvoltage[n_points];
bool state = 1;
bool previous_state = 1;
bool is_changed = 0;
unsigned long timer;


void setup() {
  Serial.begin(9600);
  //Serial.println("FIRST I PARK MY CAR.........");
  if (!driver.init()) {
    Serial.println("init failed");
  }
  
  for(int i = 0; i < n_points; i++) {
    val = analogRead(sensorpin);
    oldvoltage[i] = 0;
  }
}

void loop() {
  timer = millis();
  
  if (timer%10 == 0) {
    val = analogRead(sensorpin);

    float voltage = 5.0 / 1024 * val;
    uint8_t j;
    for(j = 1; j < n_points; j++) {
      oldvoltage[j-1] = oldvoltage[j];
    }
    oldvoltage[n_points-1] = voltage;
    for(j = 0; j < n_points; j++) {
      voltage += oldvoltage[j];
    }

    voltage /= (n_points + 1);
    
    state = (voltage >= 0.31) && (voltage <= 1.0);
    struct _packet {
      uint32_t identifier;
      uint8_t occupied;
      uint8_t battery_percentage;
    } packetdata;
  
    packetdata.identifier = 0x69696969;
    packetdata.occupied = state;
    packetdata.battery_percentage = 0x69;
  
    is_changed = !(previous_state == state);
  
    //Serial.println(voltage);
    
    if (is_changed) {
        driver.send((uint8_t*)&packetdata,sizeof(struct _packet));
        Serial.println("SENT");
        if (state) {
          //Serial.println("PARKED");
        } else {
          //Serial.println("EMPTY");
        }
        //Serial.println(voltage);
        previous_state = state;
    } 
    if (timer%30000 == 0) {
        driver.send((uint8_t*)&packetdata,sizeof(struct _packet));
        Serial.println("SENT EVERY 30");
        delay(1);
        //Serial.println(timer/1000);
    }
  }
}
