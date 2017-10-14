void setup() {
    // start serial port at 9600 bps:
    Serial.begin(9600);
    while (!Serial) {
        ; // wait for serial port to connect. Needed for native USB port only
    }
}

void loop() {
    // if we get a valid byte, read analog ins:
    //if (Serial.available() > 0) {
      // get incoming byte:
      //inByte = Serial.read();
    Serial.write("Hello, World!\n");
    delay(1000);
}