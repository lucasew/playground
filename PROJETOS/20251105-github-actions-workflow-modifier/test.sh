#!/bin/bash

set -e

echo "🧪 Testando o conversor de workflow inputs..."
echo ""

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test 1: Conversão básica
echo -e "${BLUE}Test 1: Conversão básica${NC}"
cp autorelease.yml test-output-1.yml
./convert_to_choice.py test-output-1.yml
echo -e "${GREEN}✓ Test 1 passou${NC}"
echo ""

# Test 2: Conversão com output diferente
echo -e "${BLUE}Test 2: Conversão com output diferente${NC}"
./convert_to_choice.py autorelease.yml -o test-output-2.yml
echo -e "${GREEN}✓ Test 2 passou${NC}"
echo ""

# Test 3: Modo verbose
echo -e "${BLUE}Test 3: Modo verbose${NC}"
cp autorelease.yml test-output-3.yml
./convert_to_choice.py test-output-3.yml -v
echo -e "${GREEN}✓ Test 3 passou${NC}"
echo ""

# Test 4: Workflow real
echo -e "${BLUE}Test 4: Workflow real${NC}"
./convert_to_choice.py real-autorelease.yaml -o test-output-4.yaml
echo -e "${GREEN}✓ Test 4 passou${NC}"
echo ""

# Validação
echo -e "${BLUE}Validação dos resultados${NC}"

for file in test-output-*.y*ml; do
    if grep -q "type: choice" "$file"; then
        echo -e "${GREEN}✓ $file contém 'type: choice'${NC}"
    else
        echo "✗ $file NÃO contém 'type: choice'"
        exit 1
    fi

    if grep -q "options:" "$file"; then
        echo -e "${GREEN}✓ $file contém 'options:'${NC}"
    else
        echo "✗ $file NÃO contém 'options:'"
        exit 1
    fi
done

echo ""
echo -e "${GREEN}✨ Todos os testes passaram!${NC}"

# Cleanup
echo ""
echo "🧹 Limpando arquivos de teste..."
rm -f test-output-*.y*ml
echo "✓ Limpeza concluída"
