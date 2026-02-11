package com.taskirx.sdk.config

/**
 * Configuration for AdxSDK
 */
data class AdxConfig(
    /**
     * Base API endpoint (default: production)
     */
    val apiEndpoint: String = "https://api.taskirx.com",
    
    /**
     * Enable debug logging
     */
    var enableDebug: Boolean = false,
    
    /**
     * Connection timeout in seconds
     */
    val connectionTimeout: Long = 10,
    
    /**
     * Read timeout in seconds
     */
    val readTimeout: Long = 10,
    
    /**
     * Enable test mode (uses test ads)
     */
    val testMode: Boolean = false,
    
    /**
     * Maximum ad request retries
     */
    val maxRetries: Int = 3,
    
    /**
     * Enable location-based targeting
     */
    val enableLocation: Boolean = false,
    
    /**
     * Cache ads for offline viewing
     */
    val enableAdCaching: Boolean = true,
    
    /**
     * Maximum cached ads per placement
     */
    val maxCachedAds: Int = 3,
    
    /**
     * Auto-refresh banner ads (in seconds, 0 to disable)
     */
    val bannerRefreshInterval: Int = 30,
    
    /**
     * Video ad timeout (in seconds)
     */
    val videoTimeout: Int = 30,
    
    /**
     * Enable GDPR consent dialog
     */
    val enableGDPRConsent: Boolean = true,
    
    /**
     * Enable COPPA compliance mode
     */
    val enableCOPPA: Boolean = false
)

/**
 * Ad size constants
 */
enum class AdSize(val width: Int, val height: Int) {
    BANNER_320x50(320, 50),
    BANNER_320x100(320, 100),
    BANNER_300x250(300, 250),
    BANNER_728x90(728, 90),
    LEADERBOARD_728x90(728, 90),
    MEDIUM_RECTANGLE_300x250(300, 250),
    FULL_BANNER_468x60(468, 60),
    LARGE_BANNER_320x100(320, 100);
    
    override fun toString(): String = "${width}x${height}"
}

/**
 * Ad format types
 */
enum class AdFormat {
    BANNER,
    INTERSTITIAL,
    NATIVE,
    VIDEO,
    REWARDED_VIDEO,
    IN_APP_BANNER,
    IN_APP_NATIVE,
    IN_APP_VIDEO
}

/**
 * Pricing models
 */
enum class PricingModel {
    CPM,    // Cost per mille (1000 impressions)
    CPC,    // Cost per click
    CPA,    // Cost per action
    CPI,    // Cost per install
    CPV,    // Cost per view
    CPS,    // Cost per sale
    CPR,    // Cost per registration
    DYNAMIC // Dynamic pricing
}

/**
 * Gender for targeting
 */
enum class Gender {
    MALE,
    FEMALE,
    OTHER,
    UNKNOWN
}

/**
 * Device type
 */
enum class DeviceType(val value: Int) {
    UNKNOWN(0),
    MOBILE(1),
    TABLET(5),
    DESKTOP(2),
    TV(3)
}
