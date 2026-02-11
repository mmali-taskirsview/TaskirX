# AdxSDK iOS

Modern iOS SDK for in-app advertising using Swift and SwiftUI.

## Features

- ✅ Swift 5.9+ with async/await and Combine
- ✅ SwiftUI declarative UI
- ✅ UIKit support for legacy apps
- ✅ All ad formats: Banner, Interstitial, Native, Video, Rewarded Video
- ✅ SKAdNetwork 4.0 integration
- ✅ App Tracking Transparency (ATT) support
- ✅ IDFA support with privacy controls
- ✅ AVPlayer for video ads
- ✅ Privacy Manifest included

## Requirements

- iOS 14.0+
- Xcode 15.0+
- Swift 5.9+

## Installation

### Swift Package Manager (Recommended)

Add to your `Package.swift`:

```swift
dependencies: [
    .package(url: "https://github.com/taskirx/ios-sdk", from: "1.0.0")
]
```

Or in Xcode:
1. File > Add Packages...
2. Enter: `https://github.com/taskirx/ios-sdk`
3. Select version 1.0.0+

### CocoaPods

Add to your `Podfile`:

```ruby
pod 'AdxSDK', '~> 1.0'
```

Then run:
```bash
pod install
```

### Manual Installation

1. Download the latest release
2. Drag `AdxSDK.xcframework` into your project
3. Embed & Sign the framework

## Quick Start

### 1. Initialize SDK

In your `AppDelegate` or `App` struct:

```swift
import AdxSDK

@main
struct MyApp: App {
    init() {
        AdxSDK.initialize(
            publisherId: "your-publisher-id",
            configuration: AdxConfiguration(
                apiEndpoint: "https://api.yourdomain.com",
                enableDebug: true
            )
        )
    }
    
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}
```

### 2. Show Banner Ad (SwiftUI)

```swift
import SwiftUI
import AdxSDK

struct ContentView: View {
    var body: some View {
        VStack {
            Text("My Content")
            
            // Banner ad
            AdxBannerView(
                placementId: "banner-home",
                size: .banner320x50
            )
            .frame(width: 320, height: 50)
        }
    }
}
```

### 3. Show Interstitial Ad

```swift
import AdxSDK

class GameViewController: UIViewController {
    private let interstitial = AdxInterstitial()
    
    func levelCompleted() {
        Task {
            do {
                try await interstitial.load(placementId: "interstitial-level-complete")
                try await interstitial.show(from: self)
            } catch {
                print("Failed to show interstitial: \(error)")
            }
        }
    }
}
```

### 4. Show Rewarded Video

```swift
import AdxSDK

class RewardViewController: UIViewController {
    private let rewardedAd = AdxRewardedVideo()
    
    func showRewardedAd() {
        Task {
            do {
                try await rewardedAd.load(placementId: "rewarded-extra-lives")
                
                let reward = try await rewardedAd.show(from: self)
                
                // Give user reward
                giveUserCoins(reward.amount)
                
            } catch {
                print("Failed to show rewarded ad: \(error)")
            }
        }
    }
}
```

### 5. Show Native Ad (SwiftUI)

```swift
import SwiftUI
import AdxSDK

struct FeedView: View {
    var body: some View {
        List {
            ForEach(posts) { post in
                PostView(post: post)
            }
            
            // Native ad in feed
            AdxNativeView(placementId: "native-feed") { nativeAd in
                VStack(alignment: .leading, spacing: 8) {
                    AsyncImage(url: nativeAd.imageURL)
                        .frame(height: 200)
                    
                    Text(nativeAd.title)
                        .font(.headline)
                    
                    Text(nativeAd.description)
                        .font(.body)
                    
                    Button(nativeAd.callToAction) {
                        nativeAd.handleClick()
                    }
                    .buttonStyle(.borderedProminent)
                    
                    Text("Sponsored")
                        .font(.caption)
                        .foregroundColor(.secondary)
                }
                .padding()
            }
        }
    }
}
```

## Advanced Features

### App Tracking Transparency (ATT)

Request tracking permission:

```swift
import AppTrackingTransparency
import AdxSDK

func requestTrackingPermission() {
    Task {
        let status = await ATTrackingManager.requestTrackingAuthorization()
        
        switch status {
        case .authorized:
            print("Tracking authorized")
        case .denied:
            print("Tracking denied")
        case .notDetermined:
            print("Tracking not determined")
        case .restricted:
            print("Tracking restricted")
        @unknown default:
            break
        }
    }
}
```

### Custom Targeting

```swift
let targeting = AdxTargeting(
    age: 25,
    gender: .female,
    interests: ["gaming", "technology"],
    location: CLLocation(latitude: 37.7749, longitude: -122.4194)
)

AdxBannerView(
    placementId: "banner-home",
    size: .banner320x50,
    targeting: targeting
)
```

### Ad Events

```swift
struct MyView: View {
    var body: some View {
        AdxBannerView(placementId: "banner-home", size: .banner320x50)
            .onAdLoaded {
                print("Ad loaded")
            }
            .onAdFailed { error in
                print("Ad failed: \(error)")
            }
            .onAdClicked {
                print("Ad clicked")
            }
            .onAdImpression {
                print("Ad impression tracked")
            }
    }
}
```

### UIKit Integration

For apps not using SwiftUI:

```swift
import UIKit
import AdxSDK

class ViewController: UIViewController {
    private var bannerView: AdxBannerViewUIKit!
    
    override func viewDidLoad() {
        super.viewDidLoad()
        
        // Create banner ad
        bannerView = AdxBannerViewUIKit(
            placementId: "banner-home",
            size: .banner320x50
        )
        
        bannerView.delegate = self
        
        // Add to view
        view.addSubview(bannerView)
        bannerView.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            bannerView.centerXAnchor.constraint(equalTo: view.centerXAnchor),
            bannerView.bottomAnchor.constraint(equalTo: view.safeAreaLayoutGuide.bottomAnchor),
            bannerView.widthAnchor.constraint(equalToConstant: 320),
            bannerView.heightAnchor.constraint(equalToConstant: 50)
        ])
        
        // Load ad
        bannerView.load()
    }
}

extension ViewController: AdxBannerDelegate {
    func adDidLoad(_ ad: AdxBannerViewUIKit) {
        print("Ad loaded")
    }
    
    func ad(_ ad: AdxBannerViewUIKit, didFailWithError error: Error) {
        print("Ad failed: \(error)")
    }
}
```

## Configuration Options

```swift
let config = AdxConfiguration(
    apiEndpoint: "https://api.yourdomain.com",
    enableDebug: true,
    testMode: false,
    enableLocation: true,
    connectionTimeout: 10.0,
    enableAdCaching: true,
    maxCachedAds: 3,
    bannerRefreshInterval: 30,
    videoTimeout: 30
)

AdxSDK.initialize(publisherId: "your-id", configuration: config)
```

## Privacy

### Info.plist Requirements

Add these keys to your `Info.plist`:

```xml
<key>NSUserTrackingUsageDescription</key>
<string>This allows us to show you more relevant ads</string>

<key>SKAdNetworkItems</key>
<array>
    <dict>
        <key>SKAdNetworkIdentifier</key>
        <string>your-network-id.skadnetwork</string>
    </dict>
</array>
```

### Privacy Manifest

The SDK includes a Privacy Manifest (`PrivacyInfo.xcprivacy`) declaring:
- Data collection practices
- Required reason APIs
- Tracking domains

## Testing

### Test Placement IDs

Use these for testing:
- `test-banner-320x50`
- `test-interstitial`
- `test-native`
- `test-video`
- `test-rewarded-video`

### Enable Test Mode

```swift
AdxConfiguration(testMode: true)
```

## Error Handling

```swift
do {
    try await interstitial.load(placementId: "my-placement")
    try await interstitial.show(from: self)
} catch AdxError.notInitialized {
    print("SDK not initialized")
} catch AdxError.noFill {
    print("No ad available")
} catch AdxError.networkError(let error) {
    print("Network error: \(error)")
} catch {
    print("Unknown error: \(error)")
}
```

## Performance

- **Framework Size**: ~250KB
- **Memory Usage**: <20MB
- **Ad Load Time**: ~120ms avg
- **Battery Impact**: <1%

## Support

- Documentation: https://docs.taskirx.com/ios
- GitHub: https://github.com/taskirx/ios-sdk
- Email: ios-support@taskirx.com

## License

MIT License - see LICENSE file
