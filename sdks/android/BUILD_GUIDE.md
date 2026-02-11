# Android SDK Build Guide

## Prerequisites

- Android Studio Arctic Fox or later
- JDK 11 or later
- Android SDK API 21+ (Lollipop 5.0+)
- Gradle 7.0+

---

## Build Steps

### 1. Open Project in Android Studio

```bash
cd sdks/android
# Open this directory in Android Studio
```

### 2. Configure build.gradle.kts

The SDK is already configured with:
- Kotlin 1.9+
- Jetpack Compose
- Minimum SDK: 21 (Android 5.0)
- Target SDK: 34 (Android 14)

### 3. Build AAR Library

```bash
# Command line build
./gradlew assembleRelease

# Output: build/outputs/aar/adx-sdk-release.aar
```

### 4. Test on Device/Emulator

```bash
# Run example app
./gradlew installDebug

# Or in Android Studio:
# Run > Run 'app'
```

---

## Integration Testing

### Add to Test Project

**settings.gradle.kts:**
```kotlin
dependencyResolutionManagement {
    repositories {
        google()
        mavenCentral()
        maven { url = uri("path/to/taskirx/sdks/android/build/outputs/aar") }
    }
}
```

**build.gradle.kts:**
```kotlin
dependencies {
    implementation(files("libs/adx-sdk-release.aar"))
    
    // Required dependencies
    implementation("androidx.compose.ui:ui:1.5.4")
    implementation("androidx.compose.material3:material3:1.1.2")
    implementation("com.squareup.okhttp3:okhttp:4.12.0")
    implementation("com.google.code.gson:gson:2.10.1")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.7.3")
}
```

### Test with Backend

1. Start backend server: `http://localhost:3000`
2. For Android Emulator, use: `http://10.0.2.2:3000`
3. For physical device, use: `http://YOUR_LOCAL_IP:3000`

**Test code:**
```kotlin
// Initialize SDK
AdxSDK.initialize(
    context = applicationContext,
    config = AdxConfig(
        publisherId = "test-publisher",
        apiEndpoint = "http://10.0.2.2:3000" // Emulator
    )
)

// Test banner ad
AdxBannerView(
    placementId = "test-placement-123",
    onAdLoaded = { Log.d("AdxSDK", "Ad loaded!") },
    onAdFailed = { error -> Log.e("AdxSDK", "Error: $error") }
)
```

---

## Publishing to Maven Central

### 1. Configure Gradle Properties

**gradle.properties:**
```properties
GROUP=com.taskirx
VERSION_NAME=1.0.0
POM_ARTIFACT_ID=adx-sdk

POM_NAME=TaskirX SDK
POM_DESCRIPTION=Mobile advertising SDK with OpenRTB 2.5 support
POM_URL=https://github.com/your-org/adx-sdk-android
POM_SCM_URL=https://github.com/your-org/adx-sdk-android
POM_SCM_CONNECTION=scm:git:git://github.com/your-org/adx-sdk-android.git
POM_SCM_DEV_CONNECTION=scm:git:ssh://git@github.com/your-org/adx-sdk-android.git

POM_LICENCE_NAME=MIT License
POM_LICENCE_URL=https://opensource.org/licenses/MIT
POM_DEVELOPER_ID=your-org
POM_DEVELOPER_NAME=Your Organization

SONATYPE_USERNAME=your-username
SONATYPE_PASSWORD=your-password
```

### 2. Publish

```bash
./gradlew publish
```

---

## Troubleshooting

### Build Errors

**Kotlin version mismatch:**
```bash
# Update kotlin version in build.gradle.kts
kotlin("android") version "1.9.20"
```

**Compose errors:**
```bash
# Ensure Compose is enabled
android {
    buildFeatures {
        compose = true
    }
    composeOptions {
        kotlinCompilerExtensionVersion = "1.5.4"
    }
}
```

### Runtime Errors

**Network errors:**
- Check if backend is running
- Verify API endpoint URL
- Check network permissions in AndroidManifest.xml

**Ads not loading:**
- Check logs: `adb logcat | grep AdxSDK`
- Verify placement IDs match backend campaigns
- Test with backend health endpoint first

### Testing Commands

```bash
# Clear app data
adb shell pm clear com.example.yourapp

# View logs
adb logcat | grep "AdxSDK\|taskirx"

# Install APK
adb install -r app-debug.apk

# Check device
adb devices
```

---

## MMP Integration Testing

### Test AppsFlyer

```kotlin
// In Application.onCreate()
MmpIntegrationManager.initializeAppsFlyer(
    context = this,
    devKey = "YOUR_TEST_DEV_KEY",
    appId = "com.example.testapp"
)

// Click ad and check attribution
// Check AppsFlyer dashboard for installs
```

### Test Adjust

```kotlin
MmpIntegrationManager.initializeAdjust(
    context = this,
    appToken = "YOUR_TEST_TOKEN",
    environment = AdjustConfig.ENVIRONMENT_SANDBOX
)
```

---

## Performance Testing

```bash
# Profile app
./gradlew :app:generateDebugProfile

# Memory profiling
adb shell dumpsys meminfo com.example.yourapp

# Network monitoring
adb shell am start -a android.intent.action.VIEW \
  -d "https://chrome://inspect/#devices"
```

---

## Distribution

### Internal Testing

1. Generate signed APK in Android Studio
2. Distribute via Firebase App Distribution
3. Or upload to Google Play Console (Internal Testing)

### Production

1. Publish AAR to Maven Central
2. Developers add dependency:
```kotlin
dependencies {
    implementation("com.taskirx:adx-sdk:1.0.0")
}
```

---

## Support

For Android SDK issues:
- GitHub: https://github.com/your-org/adx-sdk-android/issues
- Email: android-support@taskirx.com
