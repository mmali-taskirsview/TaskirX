# AdxSDK ProGuard Rules

# Keep all SDK public API classes
-keep public class com.taskirx.sdk.** {
    public *;
}

# Keep AdxSDK singleton
-keep class com.taskirx.sdk.AdxSDK {
    public *;
}

# Keep ad view classes
-keep class com.taskirx.sdk.ads.** {
    public *;
}

# Keep models for JSON serialization
-keep class com.taskirx.sdk.models.** {
    *;
}

# Keep config classes
-keep class com.taskirx.sdk.config.** {
    *;
}

# Moshi
-keepclassmembers class ** {
    @com.squareup.moshi.Json <fields>;
}
-keep @com.squareup.moshi.JsonQualifier @interface *
-dontwarn org.jetbrains.annotations.**
-keep class kotlin.Metadata { *; }

# OkHttp
-dontwarn okhttp3.**
-dontwarn okio.**
-keepnames class okhttp3.internal.publicsuffix.PublicSuffixDatabase

# ExoPlayer
-keep class androidx.media3.** { *; }
-dontwarn androidx.media3.**

# Google Play Services
-keep class com.google.android.gms.** { *; }
-dontwarn com.google.android.gms.**

# Jetpack Compose
-keep class androidx.compose.** { *; }
-dontwarn androidx.compose.**
