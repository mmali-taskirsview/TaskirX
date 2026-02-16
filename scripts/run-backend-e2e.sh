#!/bin/bash
cd "$(dirname "$0")/../nestjs-backend" || exit 1
npm run test:e2e
