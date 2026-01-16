#!/bin/bash

echo "=================================================="
echo "   🚀 Token Launchpad Deployment Helper"
echo "=================================================="

# Check for Docker
if ! command -v docker &> /dev/null
then
    echo "❌ Docker could not be found. Please install Docker."
else
    echo "✅ Docker is installed."
fi

# Check for Ngrok
if command -v ngrok &> /dev/null
then
    echo "✅ Ngrok is installed."
    echo "   To expose your backend for a public demo:"
    echo "   ngrok http 8081"
else
    echo "⚠️ Ngrok is NOT installed."
    echo "   To show this demo to a remote client, install ngrok:"
    echo "   https://ngrok.com/download"
fi

echo ""
echo "--- Running Production Build ---"
echo "To run the production docker container:"
echo "  docker build -t token-launchpad ."
echo "  docker run -d -p 8081:8081 --env-file .env token-launchpad"
echo ""
echo "--- Quick Local Run ---"
echo "  make run"
echo ""
