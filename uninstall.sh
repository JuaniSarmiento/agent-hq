#!/bin/bash
set -euo pipefail

echo "🏢 Uninstalling Agent HQ..."

rm -f ~/.local/bin/agenthq
rm -f ~/.local/bin/agenthq-mcp
echo "✅ Binaries removed"

read -p "Remove agent profiles? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf ~/.claude/agents/
    echo "✅ Profiles removed"
fi

read -p "Remove database? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -f ~/.claude/agenthq.db
    echo "✅ Database removed"
fi

echo "🏢 Agent HQ uninstalled"
