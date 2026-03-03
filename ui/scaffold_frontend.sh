#!/usr/bin/env bash
# gr33n frontend scaffold — run from ~/gr33n-api parent directory
set -e

cd ~
npm create vite@latest gr33n-ui -- --template vue
cd gr33n-ui
npm install
npm install vue-router@4 pinia axios
npm install -D tailwindcss@3 postcss autoprefixer
npx tailwindcss init -p

echo "✅ Dependencies installed"
