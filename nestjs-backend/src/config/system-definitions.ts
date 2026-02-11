// System Definitions for Verticals, Formats, Markets, and Audiences

export const INDUSTRY_VERTICALS = {
  GAMING_ENTERTAINMENT: {
    label: 'Gaming & Entertainment',
    subVerticals: [
      'Mobile Games', 'PC/Console Games', 'Esports', 'Gambling/Casino',
      'Fantasy Sports', 'Game Development Tools', 'Gaming Hardware'
    ],
    entertainment: [
      'Streaming Services', 'Music', 'Podcasts', 'Movie/Theater',
      'Celebrity News', 'Anime/Manga'
    ]
  },
  FINANCE_BUSINESS: {
    label: 'Finance & Business',
    finance: [
      'Cryptocurrency', 'Traditional Banking', 'Personal Finance',
      'Insurance', 'Trading Platforms', 'Payment Processors', 'Loans/Credit'
    ],
    business: [
      'SaaS', 'Marketing Tools', 'Productivity Apps', 'E-commerce Platforms',
      'Freelance Marketplaces', 'Startup/VC Ecosystem'
    ]
  },
  ECOMMERCE_RETAIL: {
    label: 'E-commerce & Retail',
    general: ['Amazon Sellers', 'Shopify Stores', 'Dropshipping', 'Flash Sales'],
    vertical: [
      'Fashion/Apparel', 'Beauty/Cosmetics', 'Health/Wellness',
      'Home & Garden', 'Electronics', 'Baby/Kids Products',
      'Pet Supplies', 'Automotive Parts', 'Food/Grocery Delivery'
    ]
  },
  HEALTH_LIFESTYLE: {
    label: 'Health & Lifestyle',
    health: [
      'Telemedicine', 'Mental Health', 'Fitness', 'Nutrition/Diet',
      'Medical Devices', 'Healthcare Services', 'Pharmaceuticals'
    ],
    lifestyle: [
      'Travel', 'Dating', 'Food Delivery', 'Real Estate',
      'Education', 'Parenting', 'DIY/Crafts'
    ]
  },
  TECHNOLOGY_SOFTWARE: {
    label: 'Technology & Software',
    tech: [
      'Mobile Apps', 'Web Services', 'Cloud Computing', 'Cybersecurity',
      'AI/ML Tools', 'Blockchain/Web3', 'IoT Devices', 'Developer Tools'
    ],
    software: [
      'Productivity Software', 'Design Tools', 'Analytics Platforms',
      'Database Services', 'Hosting Services'
    ]
  },
  SPECIALIZED: {
    label: 'Specialized Verticals',
    niche: [
      'Legal Services', 'Religious/Spiritual', 'Political', 
      'Non-Profit/Charity', 'Government Services', 'Agriculture/Farming',
      'Construction/Industrial', 'Energy/Utilities'
    ]
  }
};

export const AD_FORMATS = {
  DISPLAY: {
    banner: {
      standardSizes: ['728x90', '300x250', '160x600', '300x600', '970x250'],
    },
    richMedia: ['Expandable Banners', 'Interstitial', 'Lightbox', 'Push-down']
  },
  VIDEO: {
    inStream: ['Pre-roll', 'Mid-roll', 'Post-roll'],
    outStream: ['In-banner', 'In-article', 'In-feed'],
    ctv: ['OTT Apps', 'Smart TV', 'Gaming Consoles', 'Streaming Devices'],
    standards: ['VAST', 'VPAID', 'VMAP']
  },
  NATIVE: {
    types: ['Content Recommendation', 'In-feed Native', 'Custom Native']
  },
  MOBILE_SPECIFIC: {
    inApp: ['Mobile Interstitials', 'Rewarded Video', 'Playable Ads', 'Offerwalls'],
    web: ['AMP Ads', 'Progressive Web App Ads'],
    special: ['AR Ads', 'Location-based', 'QR Code']
  },
  AUDIO: {
    digital: ['Music Streaming', 'Podcast Ads', 'Internet Radio', 'Audiobooks'],
    programmatic: ['DAI', 'Interactive Audio', 'Voice Assistant Ads']
  },
  EMERGING: {
    advanced: ['DCO', 'Contextual Video', 'Shoppable Ads', '360 Video', 'VR Ads', 'In-game Advertising']
  },
  PERFORMANCE: {
    highIntent: ['Push Notifications', 'Popunders/Popups', 'Redirect Traffic', 'Domain Parking', 'Browser Push', 'SMS/MMS']
  }
};

export const GEO_MARKETS = {
  TIER_1: {
    NA: ['US', 'CA'],
    EU: ['UK', 'DE', 'FR', 'NL', 'CH', 'SE', 'NO', 'DK', 'FI'],
    APAC: ['AU', 'NZ', 'JP', 'KR', 'SG', 'HK', 'TW']
  },
  TIER_2: {
    EU: ['ES', 'IT', 'IE', 'AT', 'BE', 'PL', 'CZ', 'PT'],
    ASIA: ['MY', 'TH', 'VN', 'PH', 'ID', 'IN', 'IL'],
    LATAM: ['BR', 'MX', 'AR', 'CL', 'CO', 'PE']
  },
  TIER_3: {
    ASIA: ['BD', 'PK', 'LK', 'NP', 'KH', 'LA', 'MM'],
    ME: ['AE', 'SA', 'QA', 'KW', 'OM', 'BH'],
    AFRICA: ['ZA', 'NG', 'KE', 'GH', 'EG', 'MA'],
    EU_EAST: ['RO', 'HU', 'UA', 'RU', 'TR', 'GR'],
    LATAM: ['EC', 'UY', 'PY', 'BO', 'DO']
  },
  SPECIAL: {
    HIGH_GROWTH_TECH: ['EE', 'LT', 'MT'],
    RESTRICTED: ['CN', 'IR', 'KP', 'CU'],
    MICRO: ['MC', 'LU', 'IS']
  }
};

export const AUDIENCE_TYPES = {
  DEMOGRAPHICS: {
    age: ['Gen Z (13-24)', 'Millennials (25-40)', 'Gen X (41-56)', 'Baby Boomers (57-75)', 'Silent (76+)'],
    gender: ['Male', 'Female', 'Non-binary', 'Gender-neutral'],
    income: ['Low', 'Middle', 'High', 'Ultra-high-net-worth'],
    education: ['High school', 'Some college', 'Graduate', 'Post-graduate']
  },
  PSYCHOGRAPHICS: {
    lifestyle: ['Urban/Rural', 'Homeowners/Renters', 'Parents/Child-free', 'Pet owners', 'Car owners'],
    interests: ['Tech', 'Gamers', 'Sports', 'Music', 'Books', 'Movies', 'Foodies', 'Travel', 'Fitness'],
    values: ['Environmentalists', 'Social activists', 'Luxury', 'Bargain', 'Early adopters', 'Loyalists']
  },
  BEHAVIOR: {
    online: ['Shoppers', 'Bingers', 'Social', 'News', 'Mobile-first', 'Desktop-power'],
    purchase: ['Impulse', 'Research-heavy', 'Switcher', 'Loyal', 'Price-sensitive', 'Quality-focused'],
    device: ['iPhone', 'Android', 'PC', 'Tablet', 'Smart TV']
  },
  B2B: {
    size: ['Startup', 'SMB', 'Enterprise', 'Fortune 500'],
    industry: ['Tech', 'Healthcare', 'Finance', 'Manufacturing', 'Retail', 'Education'],
    role: ['C-level', 'Manager', 'IT/Dev', 'Marketing', 'Sales', 'HR']
  },
  NICHE: {
    interest: ['Crypto', 'Stock', 'Real Estate', 'Luxury Car', 'Watch', 'Wine', 'Art', 'Jet'],
    pro: ['Doctor', 'Lawyer', 'Engineer', 'Architect', 'Teacher', 'Freelancer'],
    lifeStage: ['Student', 'Grad', 'Newlywed', 'Parent', 'Empty Nester', 'Retiree']
  },
  INTENT: {
    purchase: ['In-market', 'Researching', 'Comparing', 'Abandoned Cart', 'Previous Buyer'],
    content: ['Info', 'Entertainment', 'Education', 'Solution'],
    trigger: ['Moving', 'Marriage', 'Baby', 'New Job', 'Buying Home', 'Travel']
  }
};
