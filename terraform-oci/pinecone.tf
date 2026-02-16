resource "pinecone_index" "ad_vectors" {
  name = "ad-vectors"
  dimension = 1536
  metric = "cosine"
  spec = {
    serverless = {
      cloud = "aws"
      region = "us-east-1"
    }
  }
}

output "pinecone_host" {
  value = pinecone_index.ad_vectors.host
}