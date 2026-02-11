# iOS SDK Build Guide

## Prerequisites

- Xcode 15.0 or later
- macOS 13.0 (Ventura) or later
- Swift 5.9+
- iOS 14.0+ deployment target
- CocoaPods or Swift Package Manager

---

## Build Steps

### 1. Open in Xcode

```bash
cd sdks/ios
open AdxSDK.xcodeproj
```

### 2. Configure Project Settings

The SDK is already configured with:
- Swift 5.9+
- SwiftUI
- iOS 14.0+ deployment target
- Async/await support

### 3. Build Framework

```bash
# Build for device
xcodebuild -project AdxSDK.xcodeproj \
  -scheme AdxSDK \
  -configuration Release \
  -sdk iphoneos \
  clean build

# Build for simulator
xcodebuild -project AdxSDK.xcodeproj \
  -scheme AdxSDK \
  -configuration Release \
  -sdk iphonesimulator \
  clean build

# Create XCFramework (universal)
xcodebuild -create-xcframework \
  -framework build/Release-iphoneos/AdxSDK.framework \
  -framework build/Release-iphonesimulator/AdxSDK.framework \
  -output AdxSDK.xcframework
```

### 4. Test on Device/Simulator

```bash
# Run tests
xcodebuild test \
  -project AdxSDK.xcodeproj \
  -scheme AdxSDK \
  -sdk iphonesimulator \
  -destination 'platform=iOS Simulator,name=iPhone 15 Pro'
```

---

## Swift Package Manager Integration

### Package.swift

```swift
// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "AdxSDK",
    platforms: [
        .iOS(.v14)
    ],
    products: [
        .library(
            name: "AdxSDK",
            targets: ["AdxSDK"]
        )
    ],
    dependencies: [],
    targets: [
        .target(
            name: "AdxSDK",
            dependencies: [],
            path: "Sources"
        ),
        .testTarget(
            name: "AdxSDKTests",
            dependencies: ["AdxSDK"],
            path: "Tests"
        )
    ]
)
```

### Usage in Client App

**Package.swift:**
```swift
dependencies: [
    .package(url: "https://github.com/your-org/adx-sdk-ios.git", from: "1.0.0")
]
```

**Or in Xcode:**
1. File > Add Packages...
2. Enter repository URL
3. Select version

---

## CocoaPods Integration

### Create Podspec

**AdxSDK.podspec:**
```ruby
Pod::Spec.new do |spec|
  spec.name         = "AdxSDK"
  spec.version      = "1.0.0"
  spec.summary      = "TaskirX iOS SDK with OpenRTB 2.5 support"
  spec.description  = <<-DESC
    Mobile advertising SDK for iOS with OpenRTB 2.5 support,
    multiple ad formats, and MMP integration.
  DESC
  
  spec.homepage     = "https://github.com/your-org/adx-sdk-ios"
  spec.license      = { :type => "MIT", :file => "LICENSE" }
  spec.author       = { "Your Organization" => "ios-support@taskirx.com" }
  
  spec.platform     = :ios, "14.0"
  spec.source       = { :git => "https://github.com/your-org/adx-sdk-ios.git", :tag => "#{spec.version}" }
  
  spec.source_files = "Sources/AdxSDK/**/*.swift"
  spec.swift_version = "5.9"
  
  spec.frameworks = "SwiftUI", "Combine"
end
```

### Install in Client App

**Podfile:**
```ruby
platform :ios, '14.0'
use_frameworks!

target 'YourApp' do
  pod 'AdxSDK', '~> 1.0.0'
end
```

```bash
pod install
```

---

## Integration Testing

### Test with Backend

1. Start backend server: `http://localhost:3000`
2. For Simulator, use: `http://localhost:3000`
3. For physical device, use: `http://YOUR_LOCAL_IP:3000`

**Test code:**
```swift
import SwiftUI
import AdxSDK

@main
struct TestApp: App {
    init() {
        // Initialize SDK
        AdxSDK.shared.initialize(
            publisherId: "test-publisher",
            apiEndpoint: "http://localhost:3000"
        )
    }
    
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}

struct ContentView: View {
    var body: some View {
        VStack {
            // Test banner ad
            AdxBannerView(
                placementId: "test-placement-123",
                onAdLoaded: {
                    print("✅ Ad loaded successfully!")
                },
                onAdFailed: { error in
                    print("❌ Ad failed: \(error.localizedDescription)")
                }
            )
            .frame(height: 50)
        }
    }
}
```

---

## Publishing

### 1. Version Bump

Update version in:
- `AdxSDK.podspec`
- `Package.swift`
- Xcode project settings

### 2. Tag Release

```bash
git tag -a 1.0.0 -m "Release version 1.0.0"
git push origin 1.0.0
```

### 3. Publish to CocoaPods

```bash
# Validate podspec
pod spec lint AdxSDK.podspec

# Publish to CocoaPods Trunk
pod trunk push AdxSDK.podspec
```

### 4. Publish to Swift Package Registry

```bash
# Push to GitHub
git push origin main
git push --tags

# SPM will automatically detect the package
```

---

## Troubleshooting

### Build Errors

**Swift version mismatch:**
```bash
# Update Xcode
# Ensure Swift 5.9+ is installed
swift --version
```

**Framework not found:**
```bash
# Clean build folder
rm -rf ~/Library/Developer/Xcode/DerivedData
xcodebuild clean
```

### Runtime Errors

**Network errors:**
- Check if backend is running
- Verify API endpoint URL
- Enable App Transport Security or use HTTPS

**Info.plist configuration:**
```xml
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSAllowsArbitraryLoads</key>
    <true/>
</dict>
```

**Ads not loading:**
- Check Xcode console for errors
- Verify placement IDs match backend campaigns
- Test health endpoint first:
```swift
Task {
    let url = URL(string: "http://localhost:3000/health")!
    let (data, _) = try await URLSession.shared.data(from: url)
    print(String(data: data, encoding: .utf8) ?? "")
}
```

### Testing Commands

```bash
# List simulators
xcrun simctl list devices

# Boot simulator
xcrun simctl boot "iPhone 15 Pro"

# Install app
xcrun simctl install booted path/to/YourApp.app

# View logs
xcrun simctl spawn booted log stream --level debug
```

---

## MMP Integration Testing

### Test AppsFlyer

```swift
import AdxSDK

// In AppDelegate or App init
MmpIntegrationManager.shared.initializeAppsFlyer(
    devKey: "YOUR_TEST_DEV_KEY",
    appId: "id123456789"
)

// Click ad and check attribution
// Check AppsFlyer dashboard for installs
```

### Test Adjust

```swift
MmpIntegrationManager.shared.initializeAdjust(
    appToken: "YOUR_TEST_TOKEN",
    environment: .sandbox
)
```

### Test Attribution

```swift
// Simulate click
AdxSDK.shared.trackClick(
    campaignId: "campaign-123",
    creativeId: "creative-456"
)

// Simulate install (first launch)
// MMP will send attribution to backend
```

---

## Performance Testing

### Instruments

```bash
# Profile app with Instruments
instruments -t "Time Profiler" -D trace.trace \
  -w "iPhone 15 Pro (17.0)" \
  path/to/YourApp.app
```

### Memory Leaks

```bash
# Run with Memory Graph Debugger
# Xcode > Debug > Memory Graph Debugger
```

### Network Monitoring

```bash
# Use Network Link Conditioner
# Xcode > Debug > Simulate Location
# Settings > Developer > Network Link Conditioner
```

---

## App Store Submission

### 1. Update Privacy Manifest

**PrivacyInfo.xcprivacy:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>NSPrivacyTracking</key>
    <true/>
    <key>NSPrivacyTrackingDomains</key>
    <array>
        <string>yourdomain.com</string>
    </array>
    <key>NSPrivacyCollectedDataTypes</key>
    <array>
        <dict>
            <key>NSPrivacyCollectedDataType</key>
            <string>NSPrivacyCollectedDataTypeDeviceID</string>
            <key>NSPrivacyCollectedDataTypeLinked</key>
            <true/>
            <key>NSPrivacyCollectedDataTypePurposes</key>
            <array>
                <string>NSPrivacyCollectedDataTypePurposeAppFunctionality</string>
            </array>
        </dict>
    </array>
</dict>
</plist>
```

### 2. App Store Connect

1. Create app in App Store Connect
2. Fill in required metadata
3. Add privacy policy URL
4. Explain advertising identifier usage
5. Submit for review

---

## CI/CD with GitHub Actions

**.github/workflows/ios.yml:**
```yaml
name: iOS Build

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: macos-13
    steps:
    - uses: actions/checkout@v3
    
    - name: Select Xcode
      run: sudo xcode-select -s /Applications/Xcode_15.0.app
    
    - name: Build SDK
      run: |
        xcodebuild -project AdxSDK.xcodeproj \
          -scheme AdxSDK \
          -sdk iphonesimulator \
          -configuration Release \
          clean build
    
    - name: Run Tests
      run: |
        xcodebuild test \
          -project AdxSDK.xcodeproj \
          -scheme AdxSDK \
          -sdk iphonesimulator \
          -destination 'platform=iOS Simulator,name=iPhone 15 Pro'
```

---

## Distribution

### Internal Testing

1. Archive app in Xcode
2. Distribute via TestFlight
3. Add internal testers

### Production

1. Publish framework to CocoaPods/SPM
2. Developers add dependency:

**CocoaPods:**
```ruby
pod 'AdxSDK', '~> 1.0.0'
```

**SPM:**
```swift
.package(url: "https://github.com/your-org/adx-sdk-ios.git", from: "1.0.0")
```

---

## Support

For iOS SDK issues:
- GitHub: https://github.com/your-org/adx-sdk-ios/issues
- Email: ios-support@taskirx.com
- Stack Overflow: [adx-sdk-ios]
