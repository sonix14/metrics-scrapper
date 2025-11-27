#!/bin/bash

# Проверяем наличие .env файла ИЛИ переменной окружения
if [ ! -f .env ] && [ -z "$GITHUB_TOKEN" ]; then
  echo "❌ GITHUB_TOKEN is not set"
  echo "Please create .env file with: GITHUB_TOKEN=your_token"
  echo "Or set environment variable: export GITHUB_TOKEN=your_token"
  exit 1
fi

echo "🚀 Starting everything..."

# Если есть .env файл, используем его, иначе надеемся на переменную окружения
if [ -f .env ]; then
  echo "📁 Using .env file"
  docker compose --env-file .env up -d ms-victoria-metrics grafana
  echo "⏳ Waiting for VictoriaMetrics to start..."
  sleep 10
  echo "📊 Running metrics scrapper..."
  docker compose --env-file .env up --build metrics-scrapper
else
  echo "🔑 Using environment variable"
  docker compose up -d ms-victoria-metrics grafana
  echo "⏳ Waiting for VictoriaMetrics to start..."
  sleep 10
  echo "📊 Running metrics scrapper..."
  docker compose up --build metrics-scrapper
fi

echo ""
echo "✅ Done!"
echo "📊 VictoriaMetrics: http://localhost:8428"
echo "📈 Grafana: http://localhost:3000 (admin/admin123)"