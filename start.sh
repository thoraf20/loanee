#!/bin/bash
# Start script for the loanee service

echo "🚀 Starting Loanee API..."

# Export environment (optional)
export APP_ENV=development

# Run Air for hot reload (make sure 'air' is installed)
if ! command -v air &> /dev/null
then
    echo "❌ Air is not installed. Run: go install github.com/cosmtrek/air@latest"
    exit 1
fi

air
