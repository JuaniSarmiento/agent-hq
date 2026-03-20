#!/bin/bash
set -euo pipefail

echo "🏢 Installing Agent HQ..."
echo ""

# Get the directory where this script lives
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Check Go
if ! command -v go &>/dev/null; then
    echo "❌ Go is required but not installed."
    echo "   Install from: https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
echo "✅ Go ${GO_VERSION} found"

# Build binaries
echo ""
echo "📦 Building binaries..."
cd "$SCRIPT_DIR"
go build -o bin/agenthq ./cmd/agenthq
go build -o bin/agenthq-mcp ./cmd/agenthq-mcp
echo "✅ Built agenthq and agenthq-mcp"

# Install binaries
echo ""
echo "📥 Installing to ~/.local/bin/..."
mkdir -p ~/.local/bin
cp bin/agenthq ~/.local/bin/
cp bin/agenthq-mcp ~/.local/bin/
echo "✅ Binaries installed"

# Check PATH
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    echo "⚠️  ~/.local/bin is not in your PATH. Add this to your shell profile:"
    echo "   export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

# Copy agent profiles
echo ""
echo "👤 Installing agent profiles..."
mkdir -p ~/.claude/agents
cp -r "$SCRIPT_DIR/agents/"*.md ~/.claude/agents/ 2>/dev/null || true
PROFILE_COUNT=$(ls ~/.claude/agents/*.md 2>/dev/null | wc -l)
echo "✅ ${PROFILE_COUNT} agent profiles installed to ~/.claude/agents/"

# Initialize database
echo ""
echo "🗄️  Initializing database..."
DB_PATH="${AGENTHQ_DB:-$HOME/.claude/agenthq.db}"
if command -v sqlite3 &>/dev/null; then
    sqlite3 "$DB_PATH" < "$SCRIPT_DIR/schema.sql"
    echo "✅ Database ready at ${DB_PATH}"
else
    echo "ℹ️  sqlite3 not found — database will be created automatically by the MCP server"
fi

# Done
echo ""
echo "════════════════════════════════════════════"
echo "  🏢 Agent HQ installed successfully!"
echo "════════════════════════════════════════════"
echo ""
echo "  Usage:"
echo "    agenthq          — Launch the TUI dashboard"
echo "    agenthq-mcp      — MCP server (used by Claude Code)"
echo ""
echo "  To use with Claude Code, add to settings.json:"
echo "    or install via: claude plugin install juani/agent-hq"
echo ""
