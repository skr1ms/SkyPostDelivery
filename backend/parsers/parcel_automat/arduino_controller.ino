#define CELL_DOOR_1 2
#define CELL_DOOR_2 3
#define CELL_DOOR_3 4

#define LED_1 8
#define LED_2 9
#define LED_3 10

String inputCommand = "";
boolean commandComplete = false;

void setup() {
  Serial.begin(9600);
  
  pinMode(CELL_DOOR_1, OUTPUT);
  pinMode(CELL_DOOR_2, OUTPUT);
  pinMode(CELL_DOOR_3, OUTPUT);
  
  pinMode(LED_1, OUTPUT);
  pinMode(LED_2, OUTPUT);
  pinMode(LED_3, OUTPUT);
  
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
  else if (cmd.startsWith("close_")) {
    int cellNum = cmd.substring(6).toInt();
    closeCellDoor(cellNum);
  }
  else if (cmd.startsWith("status_")) {
    int cellNum = cmd.substring(7).toInt();
    sendCellStatus(cellNum);
  }
  else if (cmd == "cells_0") {
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
  int pin, ledPin;
  
  switch(cellNum) {
    case 1:
      pin = CELL_DOOR_1;
      ledPin = LED_1;
      break;
    case 2:
      pin = CELL_DOOR_2;
      ledPin = LED_2;
      break;
    case 3:
      pin = CELL_DOOR_3;
      ledPin = LED_3;
      break;
    default:
      Serial.println("ERROR");
      return;
  }
  
  digitalWrite(pin, HIGH);
  digitalWrite(ledPin, HIGH);
  Serial.println("OK");
}

void closeCellDoor(int cellNum) {
  int pin, ledPin;
  
  switch(cellNum) {
    case 1:
      pin = CELL_DOOR_1;
      ledPin = LED_1;
      break;
    case 2:
      pin = CELL_DOOR_2;
      ledPin = LED_2;
      break;
    case 3:
      pin = CELL_DOOR_3;
      ledPin = LED_3;
      break;
    default:
      Serial.println("ERROR");
      return;
  }
  
  digitalWrite(pin, LOW);
  digitalWrite(ledPin, LOW);
  Serial.println("OK");
}

void sendCellStatus(int cellNum) {
  int pin;
  
  switch(cellNum) {
    case 1:
      pin = CELL_DOOR_1;
      break;
    case 2:
      pin = CELL_DOOR_2;
      break;
    case 3:
      pin = CELL_DOOR_3;
      break;
    default:
      Serial.println("closed");
      return;
  }
  
  if (digitalRead(pin) == HIGH) {
    Serial.println("opened");
  } else {
    Serial.println("closed");
  }
}

void closeAllDoors() {
  digitalWrite(CELL_DOOR_1, LOW);
  digitalWrite(CELL_DOOR_2, LOW);
  digitalWrite(CELL_DOOR_3, LOW);
  
  digitalWrite(LED_1, LOW);
  digitalWrite(LED_2, LOW);
  digitalWrite(LED_3, LOW);
}

