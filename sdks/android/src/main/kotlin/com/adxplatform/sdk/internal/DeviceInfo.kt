package com.taskirx.sdk.internal

import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import android.provider.Settings
import android.util.DisplayMetrics
import android.view.WindowManager
import com.taskirx.sdk.AdxSDK
import com.taskirx.sdk.config.DeviceType
import com.google.android.gms.ads.identifier.AdvertisingIdClient
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.util.Locale

/**
 * Device information collector
 */
internal class DeviceInfo(private val context: Context) {
    
    private val windowManager = context.getSystemService(Context.WINDOW_SERVICE) as WindowManager
    private val displayMetrics = DisplayMetrics().apply {
        windowManager.defaultDisplay.getMetrics(this)
    }
    
    /**
     * Get device type (mobile, tablet, etc.)
     */
    fun getDeviceType(): DeviceType {
        val screenInches = kotlin.math.sqrt(
            (displayMetrics.widthPixels / displayMetrics.xdpi.toDouble()).let { it * it } +
            (displayMetrics.heightPixels / displayMetrics.ydpi.toDouble()).let { it * it }
        )
        
        return when {
            screenInches >= 7.0 -> DeviceType.TABLET
            else -> DeviceType.MOBILE
        }
    }
    
    /**
     * Get screen width in pixels
     */
    fun getScreenWidth(): Int = displayMetrics.widthPixels
    
    /**
     * Get screen height in pixels
     */
    fun getScreenHeight(): Int = displayMetrics.heightPixels
    
    /**
     * Get device manufacturer
     */
    fun getManufacturer(): String = Build.MANUFACTURER
    
    /**
     * Get device model
     */
    fun getModel(): String = Build.MODEL
    
    /**
     * Get OS version
     */
    fun getOsVersion(): String = Build.VERSION.RELEASE
    
    /**
     * Get device language
     */
    fun getLanguage(): String = Locale.getDefault().language
    
    /**
     * Get user agent string
     */
    fun getUserAgent(): String {
        val appName = getAppName()
        val appVersion = getAppVersion()
        return "Mozilla/5.0 (Linux; Android ${Build.VERSION.RELEASE}; ${Build.MODEL}) " +
                "AppleWebKit/537.36 (KHTML, like Gecko) " +
                "Mobile/$appName/$appVersion " +
                "AdxSDK/${AdxSDK.getVersion()}"
    }
    
    /**
     * Get app name
     */
    fun getAppName(): String {
        return try {
            val appInfo = context.applicationInfo
            context.packageManager.getApplicationLabel(appInfo).toString()
        } catch (e: Exception) {
            "Unknown"
        }
    }
    
    /**
     * Get app package name (bundle ID)
     */
    fun getPackageName(): String = context.packageName
    
    /**
     * Get app version
     */
    fun getAppVersion(): String {
        return try {
            val packageInfo = context.packageManager.getPackageInfo(context.packageName, 0)
            packageInfo.versionName ?: "1.0.0"
        } catch (e: PackageManager.NameNotFoundException) {
            "1.0.0"
        }
    }
    
    /**
     * Get Google Advertising ID (GAID) asynchronously
     */
    suspend fun getAdvertisingId(): String? = withContext(Dispatchers.IO) {
        try {
            val adInfo = AdvertisingIdClient.getAdvertisingIdInfo(context)
            if (adInfo.isLimitAdTrackingEnabled) {
                AdxSDK.log("Limit Ad Tracking is enabled")
                return@withContext null
            }
            adInfo.id
        } catch (e: Exception) {
            AdxSDK.logError("Failed to get Advertising ID: ${e.message}", e)
            null
        }
    }
    
    /**
     * Check if limit ad tracking is enabled
     */
    suspend fun isLimitAdTrackingEnabled(): Boolean = withContext(Dispatchers.IO) {
        try {
            AdvertisingIdClient.getAdvertisingIdInfo(context).isLimitAdTrackingEnabled
        } catch (e: Exception) {
            false
        }
    }
    
    /**
     * Get Android ID (fallback identifier)
     */
    fun getAndroidId(): String {
        return Settings.Secure.getString(
            context.contentResolver,
            Settings.Secure.ANDROID_ID
        ) ?: "unknown"
    }
    
    /**
     * Get connection type
     */
    fun getConnectionType(): Int {
        // 0 = unknown, 2 = wifi, 3 = cellular
        // TODO: Implement network type detection
        return 0
    }
}
