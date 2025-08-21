# goexpert-otel
Exercise Observability and OpenTelemetry for postgraduate Go Expert

## Description

Desenvolver um sistema em Go que receba um CEP, identifica a cidade e retorna o clima atual (temperatura em graus celsius, fahrenheit e kelvin) juntamente com a cidade. Esse sistema deverá implementar OTEL(Open Telemetry) e Zipkin.

Basedo no cenário conhecido "Sistema de temperatura por CEP" denominado Serviço B, será incluso um novo projeto, denominado Serviço A.

### Requisitos - Serviço A (responsável pelo input):

- O sistema deve receber um input de 8 dígitos via POST, através do schema:  { "cep": "29902555" }
- O sistema deve validar se o input é valido (contem 8 dígitos) e é uma STRING
    - Caso seja válido, será encaminhado para o Serviço B via HTTP
    - Caso não seja válido, deve retornar:
        - Código HTTP: 422
        - Mensagem: invalid zipcode

### Dicas

- Para dúvidas da implementação do OTEL: https://opentelemetry.io/docs/languages/go/getting-started/
- Para implementação de spans: https://opentelemetry.io/docs/languages/go/instrumentation/#creating-spans
- Você precisará utilizar um serviço de collector do OTEL: https://opentelemetry.io/docs/collector/quick-start/
- Para mais informações sobre Zipkin: https://zipkin.io/

### Entrega:

- O código-fonte completo da implementação.
- Documentação explicando como rodar o projeto em ambiente dev.
- Utilize docker/docker-compose para que possamos realizar os testes de sua aplicação.
