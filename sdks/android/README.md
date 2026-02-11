# TaskirX Android SDK

Professional Kotlin SDK for the TaskirX advertising platform. Build, manage, and optimize ad campaigns with full type safety and coroutine support.

## Features

✨ **Complete API Coverage**
- Campaign management (CRUD + pause/resume)
- Real-time analytics and reporting
- Intelligent bidding engine with AI recommendations
- Ad placement management
- Webhook subscriptions and events
- User authentication and profile management

🚀 **Production Ready**
- Full Kotlin/Coroutine support
- Retrofit2 HTTP client with OkHttp3
- Type-safe API interfaces
- Exponential backoff retry logic
- Request timeout handling
- Comprehensive logging and debug mode
- 50+ unit and integration tests

⚡ **Performance Optimized**
- Minimal dependencies
- Efficient coroutine-based async
- Connection pooling via OkHttp
- Automatic token management
- Request deduplication ready

## Installation

### Add to build.gradle

```gradle
dependencies {
    // TaskirX SDK
    implementation 'com.taskir:android-sdk:1.0.0'
    
    // Required dependencies
    implementation 'com.squareup.retrofit2:retrofit:2.9.0'
    implementation 'com.squareup.retrofit2:converter-gson:2.9.0'
    implementation 'com.squareup.okhttp3:okhttp:4.9.3'
    implementation 'org.jetbrains.kotlinx:kotlinx-coroutines-android:1.6.4'
}
```

## Quick Start

### Initialize the SDK

```kotlin
import android.content.Context
import com.taskir.sdk.TaskirXClient
import com.taskir.sdk.data.models.ClientConfig

// Initialize the client
val config = ClientConfig(
    apiUrl = "https://api.taskir.io",
    apiKey = "your-api-key",
    debug = true
)

val client = TaskirXClient.create(context, config)
                apiEndpoint = "https://api.yourdomain.com",
                enableDebug = BuildConfig.DEBUG
            )
        )
    }
}
```

### 2. Show Banner Ad

```kotlin
// In your Activity or Fragment
@Composable
fun MyScreen() {
    Column {
        Text("My Content")
        
        // Banner ad with Jetpack Compose
        AdxBannerView(
            placementId = "banner-home",
            adSize = AdSize.BANNER_320x50,
            onAdLoaded = { Log.d("Ad", "Banner loaded") },
            onAdFailed = { error -> Log.e("Ad", "Failed: $error") }
        )
    }
}

// Or with traditional Views
val bannerView = AdxBannerView(context)
bannerView.load(
    placementId = "banner-home",
    adSize = AdSize.BANNER_320x50
)
container.addView(bannerView)
```

### 3. Show Interstitial Ad

```kotlin
val interstitial = AdxInterstitial(context)

// Load ad
interstitial.load(
    placementId = "interstitial-level-complete",
    onAdLoaded = {
        // Ad ready to show
        interstitial.show()
    },
    onAdFailed = { error ->
        Log.e("Ad", "Failed to load: $error")
    }
)
```

### 4. Show Rewarded Video Ad

```kotlin
val rewardedAd = AdxRewardedVideo(context)

rewardedAd.load(
    placementId = "rewarded-extra-lives",
    onAdLoaded = { rewardedAd.show() },
    onRewarded = { reward ->
        // User earned reward
        giveReward(reward.amount)
    },
    onAdClosed = {
        // Ad closed
    }
)
```

### 5. Show Native Ad

```kotlin
@Composable
fun NativeAdCard() {
    AdxNativeAd(
        placementId = "native-feed",
        onAdLoaded = { nativeAd ->
            NativeAdContent(nativeAd)
        }
    )
}

@Composable
fun NativeAdContent(ad: AdxNativeAdData) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Column(modifier = Modifier.padding(16.dp)) {
            AsyncImage(
                model = ad.imageUrl,
                contentDescription = "Ad image"
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(ad.title, style = MaterialTheme.typography.titleMedium)
            Text(ad.description, style = MaterialTheme.typography.bodyMedium)
            
            Button(
                onClick = { ad.performClick() }
            ) {
                Text(ad.callToAction)
            }
        }
    }
}
```

## Advanced Features

### Custom Targeting

```kotlin
val targeting = AdxTargeting(
    age = 25,
    gender = Gender.FEMALE,
    interests = listOf("gaming", "tech"),
    location = Location(lat = 37.7749, lon = -122.4194)
)

bannerView.setTargeting(targeting)
```

### Ad Events

```kotlin
val adListener = object : AdxAdListener {
    override fun onAdLoaded() {
        Log.d("Ad", "Loaded")
    }
    
    override fun onAdImpression() {
        Log.d("Ad", "Impression tracked")
    }
    
    override fun onAdClicked() {
        Log.d("Ad", "User clicked")
    }
    
    override fun onAdClosed() {
        Log.d("Ad", "Ad closed")
    }
    
    override fun onAdFailed(error: AdxError) {
        Log.e("Ad", "Failed: ${error.message}")
    }
}

bannerView.setAdListener(adListener)
```

## Architecture

The SDK uses modern Android architecture:

- **Kotlin Coroutines** for async operations
- **Flow** for reactive data streams  
- **Jetpack Compose** for declarative UI
- **OkHttp** for network requests
- **WorkManager** for background sync
- **Room** (optional) for caching

## Permissions

Add to AndroidManifest.xml:

```xml
<manifest>
    <!-- Required -->
    <uses-permission android:name="android.permission.INTERNET" />
    <uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
    
    <!-- Optional - for better targeting -->
    <uses-permission android:name="android.permission.ACCESS_COARSE_LOCATION" />
    
    <!-- For Google Play Services -->
    <uses-permission android:name="com.google.android.gms.permission.AD_ID"/>
</manifest>
```

## ProGuard Rules

```proguard
-keep class com.taskirx.sdk.** { *; }
-keepclassmembers class com.taskirx.sdk.** { *; }
```

## Testing

See example app in `/examples/android` directory.

## Support

- Documentation: https://docs.taskirx.com
- GitHub: https://github.com/taskirx/android-sdk
- Email: support@taskirx.com
