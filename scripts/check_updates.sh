#!/bin/bash
set -e

echo "Checking for outdated dependencies..."
# Get list of updates
UPDATES=$(go list -u -m -f '{{if .Update}}{{.Path}} {{.Version}} -> {{.Update.Version}}{{end}}' all)

if [[ -z "$UPDATES" ]]; then
    echo "All dependencies are up to date."
    exit 0
fi

echo "Found updates:"
echo "$UPDATES" | while read -r line; do
    MODULE=$(echo "$line" | awk '{print $1}')
    VERSION_INFO=$(echo "$line" | cut -d' ' -f2-)
    
    echo ""
    echo "📦 $MODULE ($VERSION_INFO)"
    
    # Check why it is required
    WHY=$(go mod why -m "$MODULE" 2>/dev/null | grep -v "^#")
    
    if [[ -z "$WHY" ]]; then
        echo "   (Not required by main module)"
    else
        # Indent the output
        echo "$WHY" | sed 's/^/   └─ /'
    fi
done
