# AI Agents Service Layer

This directory contains the Python-based AI microservices for TaskirX.

## Services

### 1. Ad Matching Service (`ad-matching-service`)
- **Port**: 8001 (Mapped to internal 8000)
- **Tech Stack**: FastAPI, Scikit-Learn, NumPy
- **Algorithm**: Hybrid filtering (Content-Based TF-IDF + Collaborative Filtering + Performance)
- **Endpoints**:
  - `POST /match`: Returns ranked ads for a given user profile.

## Running the Services

Services are managed via the main `docker-compose.yml`.

```bash
docker-compose up -d ad-matching
```

## Development

Each service has its own `requirements.txt`.

```bash
cd ad-matching-service
pip install -r requirements.txt
python -m uvicorn app.main:app --reload
```
