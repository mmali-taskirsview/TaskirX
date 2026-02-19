# TaskirX Supported Integrations

TaskirX is built as a polyglot, highly integratable Ad Exchange platform. Below is the comprehensive list of supported standards, protocols, and third-party services.

## 1. Ad Tech Standards & Bidding Protocols

| Protocol | Role | Description |
| :--- | :--- | :--- |
| **OpenRTB 2.5 / 2.6** | SSP & DSP | Full support for OpenRTB bid requests and responses. <br> - **SSP Endpoint**: `/openrtb` (New! Standard endpoint) <br> - **Legacy Endpoint**: `/api/ssp/auction` |
| **Native Ads 1.2** | Native Integration | Compliant with OpenRTB Native 1.2 Markup. <br> - Asset-based dynamic response generation. |
| **Prebid.js** | Header Bidding | Native adapter support for client-side header bidding. <br> - **Adapter**: `taskirxBidAdapter.js` (Prebid.org compliant) <br> - **Server Adapter**: `/api/ssp/demand-partners/templates` |
| **VAST 4.0** | Video Ad Serving | XML response generation for Video and Audio ads. <br> - Supports Linear Video, Rewarded Video. |
| **DAAST** | Audio Ad Serving | Digital Audio Ad Serving Template support. |
| **IAB TCF 2.2** | Privacy | Transparency and Consent Framework v2.2 integration endpoints for GDPR compliance. |

## 2. Mobile Measurement Partners (MMP)

Unified endpoint for mobile attribution postbacks: `POST /api/mmp/postback`

| Provider | Status | Integration Type |
| :--- | :--- | :--- |
| **AppsFlyer** | ✅ Active | Server-to-Server Postbacks |
| **Adjust** | ✅ Active | Server-to-Server Postbacks |
| **Branch** | ✅ Active | Server-to-Server Postbacks |
| **Kochava** | ✅ Active | Server-to-Server Postbacks |
| **Generic S2S** | ✅ Active | Customizable query/body parameters for any tracker |

## 3. Payment & Billing

| Service | Features |
| :--- | :--- |
| **Stripe** | - Subscription Management (Starter, Professional, Enterprise) <br> - Invoice Generation <br> - Payment Processing & Webhooks |
| **Internal Wallet** | - Balance Management <br> - Deposit/Withdrawal Ledger |

## 4. Cloud & Infrastructure

| Service | Usage |
| :--- | :--- |
| **Oracle Cloud (OCI)** | Production Environment (OKE - Kubernetes Engine, Object Storage). |
| **Prometheus** | Real-time metrics collection (`/metrics`). |
| **Grafana** | Visualization dashboards ("Polyglot Overview", "RTB Overview"). |
| **ClickHouse** | Analytics Data Warehouse (High-volume event logs). |
| **Redis** | High-speed Caching, Pub/Sub, Real-time Stats. |
| **PostgreSQL** | Primary Relational Database (User, Campaign, Billing). |

## 5. Identity & Privacy

| Standard | Status | Endpoint |
| :--- | :--- | :--- |
| **Cookie Sync** | ✅ Active | `/api/integrations/identity/sync` |
| **Unified ID 2.0 (UID2)** | ⚠️ Config Stub | Ready for API Key configuration |
| **ID5** | ⚠️ Config Stub | Ready for API Key configuration |
| **LiveRamp** | ⚠️ Config Stub | Ready for IdentityLink configuration |

## 6. Communication

| Service | Usage |
| :--- | :--- |
| **SendGrid** | Primary transactional email service. |
| **Nodemailer** | Fallback SMTP transporter. |

## 7. AI & Machine Learning

| Technology | Usage |
| :--- | :--- |
| **Scikit-Learn** | Ad Matching (TF-IDF, Logistic Regression) in Python microservices. |
| **TensorFlow.js** | Client-side/Node.js inference support. |
| **Pinecone** | Vector Database for semantic targeting and AI Agents. |

## 8. Third-Party Ad Networks (Configuration Ready)

The platform includes configuration stubs and endpoints ready to be connected with:

- **Google Ad Manager / AdX**
- **Facebook Audience Network**
- **Amazon TAM (Transparent Ad Marketplace)**
