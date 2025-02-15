# PoC compilação de app android baseado no Fava

Status: ainda não funciona

Usando python4android

## Método
- Baixa o repo do python4android: https://github.com/kivy/python-for-android
- `docker build -t p4a .` no repositório
- `docker run -ti -v $PWD:/data p4a`
  - Ativa o venv que ele criou
  - `pip install buildozer`
  - `buildozer android release`
- Vê se builda (aqui não buildou)
- Vê se executa (nem cheguei nessa parte)
