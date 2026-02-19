import pandas as pd
import numpy as np
import joblib
import logging
from sklearn.ensemble import RandomForestClassifier
from sklearn.preprocessing import StandardScaler
from collections import deque

logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO)

# Define feature columns matching `services/fraud_detector.py`
FEATURES = [
    'ip_score',         # External IP Score (mocked)
    'device_suspicion', # Logic based on device type
    'geo_distance',     # Distance change (mocked)
    'click_velocity',   # Clicks per minute
    'conversion_rate',  # Clicks to conversion
    'time_on_site',     # Seconds
    'bot_likelihood'    # From user-agent heuristics
]

def generate_synthetic_data(n_samples=10000):
    """Generate mock data for training fraud detection model"""
    logger.info(f"Generating {n_samples} samples...")
    
    # 1. Clean Traffic (98%)
    n_clean = int(n_samples * 0.98)
    clean_data = pd.DataFrame({
        'ip_score': np.random.normal(0.1, 0.05, n_clean),
        'device_suspicion': np.random.choice([0, 0.1], n_clean, p=[0.9, 0.1]),
        'geo_distance': np.random.exponential(50, n_clean), # Most users don't move far
        'click_velocity': np.random.poisson(2, n_clean),    # Normal browsing
        'conversion_rate': np.random.beta(2, 50, n_clean),  # Low conversion rate is normal
        'time_on_site': np.random.lognormal(4, 1, n_clean), # ~50s avg
        'bot_likelihood': np.random.beta(1, 20, n_clean),   # Very low
        'is_fraud': 0
    })

    # 2. Key Fraud Pattern (2%)
    n_fraud = n_samples - n_clean
    fraud_data = pd.DataFrame({
        'ip_score': np.random.normal(0.8, 0.1, n_fraud),    # High risk IP
        'device_suspicion': np.random.choice([0.8, 1.0], n_fraud),
        'geo_distance': np.random.exponential(5000, n_fraud), # Impossible travel
        'click_velocity': np.random.poisson(20, n_fraud),   # High click rate
        'conversion_rate': np.random.beta(0.1, 0.1, n_fraud), # Extremes (0 or 1)
        'time_on_site': np.random.exponential(0.5, n_fraud), # <1s duration
        'bot_likelihood': np.random.beta(10, 2, n_fraud),   # High bot score
        'is_fraud': 1
    })

    # Combine and Shuffle
    data = pd.concat([clean_data, fraud_data])
    data = data.sample(frac=1).reset_index(drop=True)
    
    # Clip values to realistic ranges
    data['ip_score'] = data['ip_score'].clip(0, 1)
    data['bot_likelihood'] = data['bot_likelihood'].clip(0, 1)
    
    return data

def train_model():
    data = generate_synthetic_data()
    
    X = data[FEATURES]
    y = data['is_fraud']
    
    # Scale Features
    scaler = StandardScaler()
    X_scaled = scaler.fit_transform(X)
    
    # Train Model
    logger.info("Training Random Forest Classifier...")
    clf = RandomForestClassifier(
        n_estimators=100,
        max_depth=5,
        random_state=42,
        class_weight='balanced'
    )
    clf.fit(X_scaled, y)
    
    # Evaluate
    score = clf.score(X_scaled, y)
    logger.info(f"Model Accuracy: {score:.4f}")
    
    # Feature Importance
    importances = dict(zip(FEATURES, clf.feature_importances_))
    logger.info(f"Feature Importances: {importances}")
    
    # Save artifacts
    output_dir = "app/models"
    os.makedirs(output_dir, exist_ok=True)
    
    model_path = os.path.join(output_dir, "fraud_detector.pkl")
    scaler_path = os.path.join(output_dir, "scaler.pkl")
    
    joblib.dump(clf, model_path)
    joblib.dump(scaler, scaler_path)
    logger.info(f"Model saved to {model_path}")

if __name__ == "__main__":
    import os
    train_model()
