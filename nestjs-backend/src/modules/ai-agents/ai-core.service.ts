import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import * as tf from '@tensorflow/tfjs';

@Injectable()
export class AiCoreService implements OnModuleInit {
  private readonly logger = new Logger(AiCoreService.name);
  
  private bidModel: tf.Sequential;
  private fraudModel: tf.LayersModel;
  private predictionModel: tf.Sequential;

  async onModuleInit() {
    this.logger.log('Initializing AI Models...');
    await this.initBidModel();
    await this.initFraudModel();
    await this.initPredictionModel();
    this.logger.log('AI Models initialized successfully');
  }

  // --- 1. Bid Optimization Model (Deep Neural Network) ---
  private async initBidModel() {
    this.bidModel = tf.sequential();
    
    // Input: 10 features (campaign metrics, context, etc.)
    // Layers: 64 -> 32 -> 16 -> 1 (Sigmoid for normalized bid factor)
    this.bidModel.add(tf.layers.dense({ units: 64, activation: 'relu', inputShape: [10] }));
    this.bidModel.add(tf.layers.dropout({ rate: 0.2 }));
    this.bidModel.add(tf.layers.dense({ units: 32, activation: 'relu' }));
    this.bidModel.add(tf.layers.dense({ units: 16, activation: 'relu' }));
    this.bidModel.add(tf.layers.dense({ units: 1, activation: 'sigmoid' }));

    this.bidModel.compile({ optimizer: 'adam', loss: 'meanSquaredError' });
    
    // Warm-up with dummy data to build weights
    const dummyInput = tf.zeros([1, 10]);
    this.bidModel.predict(dummyInput);
    dummyInput.dispose();
  }

  async predictOptimalBid(features: number[]): Promise<{ bidMultiplier: number; confidence: number }> {
    return tf.tidy(() => {
      const input = tf.tensor2d([features], [1, 10]);
      const prediction = this.bidModel.predict(input) as tf.Tensor;
      const bidMultiplier = prediction.dataSync()[0];
      
      // Simulated confidence based on variance or closeness to 0.5 (heuristic)
      const confidence = 0.5 + (Math.abs(bidMultiplier - 0.5)); 

      return { bidMultiplier, confidence };
    });
  }

  async trainBidModel(inputs: number[][], targets: number[]) {
    const xs = tf.tensor2d(inputs, [inputs.length, 10]);
    const ys = tf.tensor2d(targets, [targets.length, 1]);

    await this.bidModel.fit(xs, ys, {
      epochs: 5,
      batchSize: 32,
      shuffle: true
    });

    xs.dispose();
    ys.dispose();
    this.logger.log(`Bid Model retrained with ${inputs.length} samples`);
  }

  // --- 2. Fraud Detection Model (Autoencoder) ---
  private async initFraudModel() {
    // 15 indicators -> Compressed to 8 -> Reconstructed to 15
    const input = tf.input({ shape: [15] });
    const encoder = tf.layers.dense({ units: 8, activation: 'relu' }).apply(input);
    const decoder = tf.layers.dense({ units: 15, activation: 'sigmoid' }).apply(encoder);
    
    this.fraudModel = tf.model({ inputs: input, outputs: decoder as tf.SymbolicTensor });
    this.fraudModel.compile({ optimizer: 'adam', loss: 'meanSquaredError' });

    // Warm-up
    const dummyInput = tf.zeros([1, 15]);
    this.fraudModel.predict(dummyInput);
    dummyInput.dispose();
  }

  async detectFraud(indicators: number[]): Promise<{ isFraud: boolean; score: number }> {
    return tf.tidy(() => {
      const input = tf.tensor2d([indicators], [1, 15]);
      const reconstructed = this.fraudModel.predict(input) as tf.Tensor;
      
      // Calculate Mean Squared Error between input and reconstruction
      const mse = tf.losses.meanSquaredError(input, reconstructed) as tf.Tensor;
      const score = mse.dataSync()[0];

      // Threshold for anomaly (this would be tuned dynamically in production)
      const threshold = 0.3; 

      return { 
        isFraud: score > threshold, 
        score 
      };
    });
  }

  // --- 3. Predictive Analytics (LSTM) ---
  private async initPredictionModel() {
    this.predictionModel = tf.sequential();
    
    // Sequence of 7 days, 5 features per day
    this.predictionModel.add(tf.layers.lstm({ 
      units: 32, 
      inputShape: [7, 5],
      returnSequences: false 
    }));
    this.predictionModel.add(tf.layers.dense({ units: 5, activation: 'linear' })); // Predict next day's 5 metrics

    this.predictionModel.compile({ optimizer: 'adam', loss: 'meanSquaredError' });
    
     // Warm-up
     const dummyInput = tf.zeros([1, 7, 5]);
     this.predictionModel.predict(dummyInput);
     dummyInput.dispose();
  }

  async predictNextDayPerformance(history7Days: number[][]): Promise<number[]> {
    // history7Days should be 7x5 array
    if (history7Days.length !== 7 || history7Days[0].length !== 5) {
        throw new Error('Invalid input shape for prediction. Expected 7x5.');
    }

    return tf.tidy(() => {
      const input = tf.tensor3d([history7Days], [1, 7, 5]);
      const prediction = this.predictionModel.predict(input) as tf.Tensor;
      return Array.from(prediction.dataSync());
    });
  }
}
