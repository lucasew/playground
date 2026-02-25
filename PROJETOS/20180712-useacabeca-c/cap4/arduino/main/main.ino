// Não tenho a placa mas vou deixar o código pelo menos compilável
#define PIN_LED 13
#define PIN_SU 0

short umidade = 0;
int led_state = LOW;

void setup() {
    pinMode(PIN_LED, OUTPUT);
    pinMode(PIN_SU, INPUT);
    Serial.begin(9600);
}

void loop() {
    umidade = analogRead(PIN_SU); // Pegar umidade
    if (umidade < 800) { // Usando o exemplo deles
        led_state = HIGH;
        Serial.println("Estou com sede!");
    } else if (umidade < 600) {
        led_state = !led_state;
        Serial.println("Estou com muita sede, muita sede mesmo :( ");
    } else {
        led_state = LOW;
        Serial.println("Obrigado, Seymour!");
    }
    digitalWrite(PIN_LED, led_state);
    sleep(1000); // Vamos com calma gente :v
}

// O compilador ficou pedindo aqui, na duvida vou deixar
int main() {
  setup();
  while (1) {
    loop();
  }
}

