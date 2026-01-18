# Lanchonete do ZAMIGOS

App criado para o minicurso _Fullstack Javascript do zero ao deploy_ da Semana Acadêmica de Ciência da Computação da UTFPR-SH.

## About

O App se trata de um sistema pedidos para lanchonetes. Nele, é possível ter a criação de comandas (orders) com o envolvimento de 2 níveis de usuários: administradores e atendentes.

Pelo Single Page Application (SPA) construído em React, será possível ficar a par das comandas que estão ativas (novos pedidos), assim como encerrá-las (pedido pronto e retirado, não por meio do pagamento).

Pelo app mobile construído em React Native, os atendentes podem realizar a criação de comandas, por meio de atendimentos na mesa. Essa será a única função dos atendentes.

A única função do administrador é criar os lanches (snacks) e bebidas (drinks).

## Tecnologias

- Backend (API)
  - Express
  - Mongoose
  - JWT (JsonWebToken)
  - dotenv
  - cors

- Frontend (React)
  - React
    - React-Router-Dom
  - Axios

- Mobile
  - React Native
    - React Navigation
  - Axios

## Consumir API via Insomnia

Utilize o arquivo `zamigos.json` e importe ao Insomnia para ter as configurações pronta para consumo.

:: Também é possível usar em outros aplicativos, como `Postman`.

### Passo a passo p/ importação

- Insominia
  1. Abra o insomnia
  2. Vá na aba Application -> Preferences -> Data -> Import Data -> `zamigos.json`
  3. Certifique-se que seu servidor backend estará ouvindo no endereço e porta `localhost:3000`

## Contato

- [Jonathan Galdino](https://github.com/jonathangaldino)
- [Rafael Boniolo](https://github.com/rafaelboniolo)
