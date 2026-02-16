import pinecone
import os
import time

# --- PINE CONE CONFIGURATION ---
API_KEY = os.getenv("PINECONE_API_KEY", "your-api-key")  # Replace or set env var
ENVIRONMENT = os.getenv("PINECONE_ENV", "us-west1-gcp-free")
INDEX_NAME = "ad-vectors"
DIMENSION = 1536  # Must match your embedding model (e.g., Ada-002: 1536)

try:
    pinecone.init(api_key=API_KEY, environment=ENVIRONMENT)
except Exception as e:
    print(f"Failed to init Pinecone: {e}")
    exit(1)

# --- 1. CREATE INDEX (IF NOT EXISTS) ---
active_indexes = pinecone.list_indexes()
if INDEX_NAME not in active_indexes:
    print(f"Creating index: {INDEX_NAME}...")
    pinecone.create_index(name=INDEX_NAME, dimension=DIMENSION, metric="cosine")
    time.sleep(30)  # Wait for index to be ready
    print("Index created!")

index = pinecone.Index(INDEX_NAME)

# --- 2. UPSERT VECTORS (AD PROFILE DATA) ---
# Format: (id, vector, metadata)
# Example 'ad1': Ad promoting running shoes
sample_vectors = [
    ("ad1", [0.1] * DIMENSION, {"type": "shoes", "target_audience": "runner"}),
    ("ad2", [0.2] * DIMENSION, {"type": "laptop", "target_audience": "developer"}),
    ("ad3", [0.3] * DIMENSION, {"type": "coffee", "target_audience": "everyone"}),
]

upsert_response = index.upsert(vectors=sample_vectors)
print(f"Upserted: {upsert_response}")

# --- 3. QUERY SIMILAR ADS (REAL-TIME BIDDING) ---
# Imagine a user visits a page about 'marathons' -> Embedding vector [0.11, ...]
user_vector = [0.11] * DIMENSION  # Simulated similar vector to 'ad1'

query_response = index.query(
    vector=user_vector,
    top_k=3,
    include_metadata=True
)

print("\n--- QUERY RESULTS (User interested in running) ---")
for match in query_response['matches']:
    print(f"Ad found: {match['id']}, Score: {match['score']:.4f}, Details: {match['metadata']}")
