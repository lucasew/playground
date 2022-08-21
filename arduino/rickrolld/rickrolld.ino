#include <WiFiServer.h>
#include <WiFiClient.h>
#include <WiFiUdp.h>
#include <WiFi.h>

char* rickroll = "We're no strangers to love\nYou know the rules and so do I\nA full commitments what I'm thinking of\nYou wouldn't get this from any other guy\n\nI just wanna tell you how I'm feeling\nGotta make you understand\n\nNever gonna give you up\nNever gonna let you down\nNever gonna turn around and desert you\nNever gonna make you cry\nNever gonna say goodbye\nNever gonna tell a lie and hurt you\n\nWe've known each other for so long\nYour heart's been aching\nBut you're too shy to say it\nInside we both know what's been going on\nWe know the game and we're gonna play it\n\nAnd if you ask me how I'm feeling\nDon't tell me you're too blind to see\n\nNever gonna give you up\nNever gonna let you down\nNever gonna turn around and desert you\nNever gonna make you cry\nNever gonna say goodbye\nNever gonna tell a lie and hurt you\n\nNever gonna give you up\nNever gonna let you down\nNever gonna turn around and desert you\nNever gonna make you cry\nNever gonna say goodbye\nNever gonna tell a lie and hurt you\n\nOoh (give you up)\nOoh (give you up)\nNever gonna give, never gonna give (ooh, give you up)\nNever gonna give, never gonna give (ooh, give you up)\n\nWe've known each other for so long\nYour heart's been aching\nBut you're too shy to say it\nInside we both know what's been going on\nWe know the game and we're gonna play it\n\nI just wanna tell you how I'm feeling\nGotta make you understand\n\nNever gonna give you up\nNever gonna let you down\nNever gonna turn around and desert you\nNever gonna make you cry\nNever gonna say goodbye\nNever gonna tell a lie and hurt you\n\nNever gonna give you up\nNever gonna let you down\nNever gonna turn around and desert you\nNever gonna make you cry\nNever gonna say goodbye\nNever gonna tell a lie and hurt you\n\nNever gonna give you up\nNever gonna let you down\nNever gonna turn around and desert you\nNever gonna make you cry\nNever gonna say goodbye\nNever gonna tell a lie and hurt you";

IPAddress me(69,69,69,1);
IPAddress gateway(69,69,69,1);
IPAddress subnet(255,255,255,0);

#define MAX_CLIENTS 20
WiFiClient *clients;
int *indexes;

WiFiServer rickrolld(420);

void setup() {
  Serial.begin(9600);
  WiFi.softAPConfig(me, gateway, subnet);
  WiFi.softAP("rickrolld");
  Serial.println(WiFi.softAPIP());
  clients = (WiFiClient*) malloc(MAX_CLIENTS*sizeof(WiFiClient));
  assert(clients);
  indexes = (int*) malloc(MAX_CLIENTS*sizeof(int));
  assert(indexes);
  memset(clients, 0, 20*sizeof(WiFiClient));
  memset(indexes, 0, 20*sizeof(int));
  Serial.println("Esperando wifi subir...");
//  while (WiFi.status() != WL_CONNECTED) {delay(100);}
  Serial.println("Wifi subiu, rickroll time!");
  rickrolld.begin();
  Serial.printf("%i", clients[0]);
  Serial.printf("%i", indexes[0]);
  Serial.flush();
  
}

void handle_connections() {
  for (int i; i < MAX_CLIENTS; i++) {
    WiFiClient client = clients[i];
    if (!client) {
      continue;
    }
    Serial.println(client);
    int index = indexes[i];
    char c = rickroll[i];
    if (c) {
      if (client.connected()) {
        client.printf("%c", c);
      } else {
        Serial.printf("client %i desconectado\n", i);
        clients[i] = NULL;
        indexes[i] = 0;
      }
    } else {
      indexes[i] = 0;
      client.flush();
    }
  }
}

void loop() {
//  Serial.println("* loop *");
  WiFiClient client = rickrolld.available();
  if (client) {
    Serial.println("novo client");
    Serial.flush();
    client.println("foi?");
    client.flush();
    client.stop();
    clients[0] = client;
//    
//    char foundSlot = 0;
//    for (int i; i < MAX_CLIENTS; i++) {
//      if (!clients[i]) {
//        Serial.printf("novo client no index %i\n", i);
//        Serial.flush();
//        clients[i] = client;
//        indexes[i] = 0;
//        foundSlot = 1;
//        break;
//      }
//      if (!foundSlot) {
//        client.println("No slots to be used");
//        client.flush();
//        client.stop();
//      }
//    }
  }
  handle_connections();
}
