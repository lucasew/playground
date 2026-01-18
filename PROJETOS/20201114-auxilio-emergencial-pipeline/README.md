# Implementação de pipeline para processamento de dados do auxilio emergencial

O pipeline foi projetado para transformar os arquivos ZIP do portal da transparência em um ou mais bancos de dados SQLite.

A implementação desse pipeline foi descrita com mais detalhes [neste post](http://localhost:1313/post/auxilio-emergencial/)

## O que cada utilitário faz?

- zipcat: cat para arquivos zip
  - recebe: [nome do arquivo zip] [outras flags do comando zip]
  - stdin: ignorado
  - stdout: dados descomprimidos, por padrão de todos os arquivos, como o zip só tem um CSV é safe.
- sqlify: transforma o CSV no shape do CSV do auxilio emergencial para o SQL do banco de dados, foi projetado para código SQLite mas pode ser facilmente adaptado
  - recebe: nenhum parâmetro
  - stdin: csv
  - stdout: sql
- sql2db: aplica o sql especificado em um banco de dados SQLite
  - recebe: caminho para o banco de dados, se o arquivo não existir ele vai ser criado
  - stdin: sql a ser executado
  - stdout: algum resultado que algum sql de repente retorna
- dbify_all_zips: recebe o nome dos zips usando aquela notação curinga, tipo `caminho/para/pasta/*` e aplica o pipeline em cada zip para um banco sqlite
  - recebe: pasta com zips a serem processados; caminho onde será salvo o banco de dados
  - stdin: nada
  - stdout: logs, relatórios, nada machine friendly mesmo :v
  - literalmente é um for que passa em cada zip e roda os outros 3 utilitários acima passando a saída de um como entrada de outro

**TODO**: Comparativo de velocidade usando diferentes tweaks no SQL.

Basicamente a tunagem fica no arquivo `sqlify.awk`. As otimizações do SQL, de loop e tudo mais. O resto é só plumbing mesmo.
