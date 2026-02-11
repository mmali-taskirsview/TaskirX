# Python Bid Optimization Service
FROM python:3.10-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt
RUN pip install prometheus-fastapi-instrumentator

# Copy application code
COPY . .

# Set environment
ENV PYTHONUNBUFFERED=1
ENV PORT=6003

# Expose port
EXPOSE 6003

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=30s --retries=3 \
  CMD curl -f http://localhost:6003/api/health || exit 1

# Run the application
CMD ["python", "-m", "app.main"]
