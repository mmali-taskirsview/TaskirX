resource "pinecone_index" "taskir_ads" {
  name      = "taskir-ads"
  dimension = 1536 # For OpenAI embeddings or similar
  metric    = "cosine"
  spec = {
    serverless = {
      cloud = "aws" 
      region = "us-east-1"
    }
  }
}

output "pinecone_index_host" {
  value = pinecone_index.taskir_ads.host
}
