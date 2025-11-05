#!/usr/bin/env -S uv run
# /// script
# dependencies = ["ruamel.yaml"]
# ///
"""
Converte inputs de workflow_dispatch para tipo choice com opções [patch, minor, major].
Salva inplace por padrão.

Uso:
    ./convert_to_choice.py <workflow.yml> [-o output.yml] [-v]
"""

import argparse
from ruamel.yaml import YAML

CHOICES = ["patch", "minor", "major"]


def main():
    parser = argparse.ArgumentParser(description='Converte inputs para choice')
    parser.add_argument('workflow', help='Arquivo YAML do workflow')
    parser.add_argument('-o', '--output', help='Arquivo de saída (padrão: inplace)')
    parser.add_argument('-v', '--verbose', action='store_true')

    args = parser.parse_args()

    yaml = YAML()
    yaml.preserve_quotes = True
    yaml.width = 4096

    print(f"📖 Lendo: {args.workflow}")
    with open(args.workflow, 'r') as f:
        data = yaml.load(f)

    if 'on' not in data:
        print("❌ Seção 'on' não encontrada")
        return

    on_section = data['on']
    if not isinstance(on_section, dict) or 'workflow_dispatch' not in on_section:
        print("❌ workflow_dispatch não encontrado")
        return

    wd = on_section['workflow_dispatch']
    if wd is None:
        wd = on_section['workflow_dispatch'] = {}

    if 'inputs' not in wd:
        wd['inputs'] = {}

    inputs = wd['inputs']
    if not inputs:
        print("⚠️  Nenhum input encontrado")
        return

    count = 0
    for name, config in inputs.items():
        if config is None:
            config = inputs[name] = {}

        input_type = config.get('type', 'string') if isinstance(config, dict) else 'string'

        if input_type in ('string', None) or 'type' not in config:
            if args.verbose:
                print(f"  ✓ Convertendo '{name}' para choice")

            config['type'] = 'choice'
            config['options'] = CHOICES

            if 'default' in config and config['default'] not in CHOICES:
                config['default'] = CHOICES[0]

            count += 1
        elif args.verbose:
            print(f"  - Mantendo '{name}' como {input_type}")

    print(f"✓ {count} input(s) convertido(s)")

    output = args.output or args.workflow
    print(f"💾 Salvando em: {output}")
    with open(output, 'w') as f:
        yaml.dump(data, f)

    print("✨ Concluído!")


if __name__ == '__main__':
    main()
