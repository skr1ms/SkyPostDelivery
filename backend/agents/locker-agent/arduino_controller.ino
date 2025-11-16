#include <Servo.h>

#define SERVO_INTERNAL_1 A1
#define SERVO_INTERNAL_2 A2
#define SERVO_INTERNAL_3 A3

#define SERVO_DOOR_1 A4
#define SERVO_DOOR_2 A5
#define SERVO_DOOR_3 A6

Servo servoInternal1;
Servo servoInternal2;
Servo servoInternal3;

Servo servoDoor1;
Servo servoDoor2;
Servo servoDoor3;

String inputCommand = "";
boolean commandComplete = false;

void setup() {
  Serial.begin(9600);
 
  servoInternal1.attach(SERVO_INTERNAL_1);
  servoInternal2.attach(SERVO_INTERNAL_2);
  servoInternal3.attach(SERVO_INTERNAL_3);
  
  servoDoor1.attach(SERVO_DOOR_1);
  servoDoor2.attach(SERVO_DOOR_2);
  servoDoor3.attach(SERVO_DOOR_3);
  
  closeAllDoors();
  
  Serial.println("OK");
}

void loop() {
  if (Serial.available() > 0) {
    char inChar = (char)Serial.read();
    
    if (inChar != '\n' && inChar != '\r') {
      inputCommand += inChar;
    }
    
    if (inChar == '\n') {
      commandComplete = true;
    }
  }
  
  if (commandComplete) {
    inputCommand.trim();
    processCommand(inputCommand);
    inputCommand = "";
    commandComplete = false;
  }
}

void processCommand(String cmd) {
  if (cmd.startsWith("open_")) {
    int cellNum = cmd.substring(5).toInt();
    openCellDoor(cellNum);
  }
  else if (cmd.startsWith("internal_")) {
    int cellNum = cmd.substring(9).toInt();
    openInternalDoor(cellNum);
  }
  else if (cmd == "cells") {
    Serial.println("3");
  }
  else if (cmd == "reset") {
    closeAllDoors();
    Serial.println("OK");
  }
  else {
    Serial.println("ERROR");
  }
}

void openCellDoor(int cellNum) {
  Servo* servo;
  
  switch(cellNum) {
    case 1:
      servo = &servoDoor1;
      break;
    case 2:
      servo = &servoDoor2;
      break;
    case 3:
      servo = &servoDoor3;
      break;
    default:
      Serial.println("ERROR: Cell Num not found");
      return;
  }
  
  smoothMove(servo, 90, 75, 200);
  delay(5*1000);
  smoothMove(servo, 75, 90, 200);

  Serial.println("OK");
}

void openInternalDoor(int cellNum) {
  Servo* servo;
  
  switch(cellNum) {
    case 1:
      servo = &servoInternal1;
      break;
    case 2:
      servo = &servoInternal2;
      break;
    case 3:
      servo = &servoInternal3;
      break;
    default:
      Serial.println("ERROR: Internal Door Num not found");
      Serial.println(cellNum);
      return;
  }

  smoothMove(servo, 45, 0, 200);
  delay(10*1000);
  smoothMove(servo, 0, 45, 200);

  Serial.println("OK");
}

void closeAllDoors() {  
  servoInternal1.write(42);
  delay(1000);
  servoInternal2.write(42);
  delay(1000);
  servoInternal3.write(42);
  delay(1000);

  servoDoor1.write(90);
  delay(1000);
  servoDoor2.write(90);
  delay(1000);
  servoDoor3.write(90);
  delay(1000);
}

void smoothMove(Servo* servo, int startAngle, int endAngle, int delayTime) {
  if (startAngle < endAngle) {
    for (int pos = startAngle; pos <= endAngle; pos++) {
      servo->write(pos);
      delay(delayTime);
    }
  } else {
    for (int pos = startAngle; pos >= endAngle; pos--) {
      servo->write(pos);
      delay(delayTime);
    }
  }
}