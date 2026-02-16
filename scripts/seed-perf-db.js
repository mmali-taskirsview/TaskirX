const { Client } = require('pg');
const fs = require('fs');
const path = require('path');

// Default config - override with env vars if needed
const config = {
  host: process.env.POSTGRES_HOST || 'localhost',
  port: process.env.POSTGRES_PORT || 5432,
  user: process.env.POSTGRES_USER || 'postgres',
  password: process.env.POSTGRES_PASSWORD || 'postgres',
  database: process.env.POSTGRES_DB || 'taskirx',
};

async function seed() {
  const client = new Client(config);
  
  try {
    console.log(`Connecting to database ${config.database} at ${config.host}:${config.port}...`);
    await client.connect();
    console.log('Connected successfully.');

    const sqlPath = path.join(__dirname, 'seed-perf-data.sql');
    if (!fs.existsSync(sqlPath)) {
      throw new Error(`Seed file not found at ${sqlPath}`);
    }

    const sql = fs.readFileSync(sqlPath, 'utf8');
    
    console.log('Executing seed script...');
    await client.query(sql);
    
    console.log('✅ Performance test data seeded successfully!');
  } catch (err) {
    console.error('❌ Error seeding database:', err.message);
    if (err.code === 'ECONNREFUSED') {
        console.error('   Verify Postgres is running and accessible.');
    }
    process.exit(1);
  } finally {
    await client.end();
  }
}

seed();
