#!/bin/sh
# Install the scaffold-go-cli agent skill into your Claude skills directory.
#
# Works two ways:
#   - From a clone:  sh install.sh           (copies the in-repo SKILL.md)
#   - Over the wire: curl -fsSL .../install.sh | sh   (downloads SKILL.md)
#
# Override the destination with CLAUDE_SKILLS_DIR.
set -eu

SKILL_NAME="scaffold-go-cli"
RAW_URL="https://raw.githubusercontent.com/NickMoignard/go_cli_tool_template/master/skills/scaffold-go-cli/SKILL.md"

DEST_ROOT="${CLAUDE_SKILLS_DIR:-$HOME/.claude/skills}"
DEST_DIR="$DEST_ROOT/$SKILL_NAME"
DEST_FILE="$DEST_DIR/SKILL.md"

# Resolve the directory this script lives in (best effort; empty when piped).
SCRIPT_DIR=""
case "$0" in
  */*) SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" 2>/dev/null && pwd) || SCRIPT_DIR="" ;;
esac

LOCAL_SKILL=""
if [ -n "$SCRIPT_DIR" ] && [ -f "$SCRIPT_DIR/skills/$SKILL_NAME/SKILL.md" ]; then
  LOCAL_SKILL="$SCRIPT_DIR/skills/$SKILL_NAME/SKILL.md"
fi

mkdir -p "$DEST_DIR"

if [ -n "$LOCAL_SKILL" ]; then
  cp "$LOCAL_SKILL" "$DEST_FILE"
  echo "Copied skill from clone: $LOCAL_SKILL"
elif command -v curl >/dev/null 2>&1; then
  curl -fsSL "$RAW_URL" -o "$DEST_FILE"
  echo "Downloaded skill via curl: $RAW_URL"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$DEST_FILE" "$RAW_URL"
  echo "Downloaded skill via wget: $RAW_URL"
else
  echo "Error: no local SKILL.md found and neither curl nor wget is available." >&2
  echo "Install curl or wget, or run this script from a clone of the repo." >&2
  exit 1
fi

if [ ! -s "$DEST_FILE" ]; then
  echo "Error: install produced an empty file at $DEST_FILE" >&2
  exit 1
fi

echo "Installed scaffold-go-cli skill to: $DEST_FILE"
echo "Now invoke /scaffold-go-cli in your agent to scaffold a new Go CLI."
