package com.taskirx.sdk

import android.app.Application
import android.content.Context
import android.util.Log
import com.taskirx.sdk.config.AdxConfig
import com.taskirx.sdk.internal.DeviceInfo
import com.taskirx.sdk.internal.NetworkClient
import com.taskirx.sdk.internal.UserIdManager
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import java.util.UUID

/**
 * Main entry point for the taskirx Android SDK
 * 
 * Initialize once in your Application class:
 * ```
 * class MyApp : Application() {
 *     override fun onCreate() {
 *         super.onCreate()
 *         AdxSDK.init(this, "your-publisher-id")
 *     }
 * }
 * ```
 */
object AdxSDK {
    
    private const val TAG = "AdxSDK"
    private const val SDK_VERSION = "1.0.0"
    
    private var isInitialized = false
    private lateinit var applicationContext: Context
    private lateinit var configuration: AdxConfig
    private lateinit var networkClient: NetworkClient
    private lateinit var deviceInfo: DeviceInfo
    private lateinit var userIdManager: UserIdManager
    
    internal val sdkScope = CoroutineScope(SupervisorJob() + Dispatchers.Main)
    
    var publisherId: String = ""
        private set
    
    var sessionId: String = UUID.randomUUID().toString()
        private set
    
    /**
     * Initialize the AdxSDK
     * 
     * @param context Application context
     * @param publisherId Your publisher ID from taskirx
     * @param config Optional configuration
     */
    fun init(
        context: Context,
        publisherId: String,
        config: AdxConfig = AdxConfig()
    ) {
        if (isInitialized) {
            log("SDK already initialized")
            return
        }
        
        require(publisherId.isNotBlank()) {
            "Publisher ID cannot be blank"
        }
        
        applicationContext = context.applicationContext
        this.publisherId = publisherId
        this.configuration = config
        
        // Initialize internal components
        deviceInfo = DeviceInfo(applicationContext)
        userIdManager = UserIdManager(applicationContext)
        networkClient = NetworkClient(config)
        
        isInitialized = true
        
        log("AdxSDK v$SDK_VERSION initialized for publisher: $publisherId")
        log("API Endpoint: ${config.apiEndpoint}")
        log("Session ID: $sessionId")
    }
    
    /**
     * Check if SDK is initialized
     */
    fun isInitialized(): Boolean = isInitialized
    
    /**
     * Get SDK version
     */
    fun getVersion(): String = SDK_VERSION
    
    /**
     * Enable or disable debug logging
     */
    fun setDebugEnabled(enabled: Boolean) {
        ensureInitialized()
        configuration.enableDebug = enabled
    }
    
    /**
     * Get the user ID (persistent across sessions)
     */
    fun getUserId(): String {
        ensureInitialized()
        return userIdManager.getUserId()
    }
    
    /**
     * Set custom user ID (optional)
     */
    fun setUserId(userId: String) {
        ensureInitialized()
        userIdManager.setCustomUserId(userId)
    }
    
    /**
     * Get Google Advertising ID (asynchronously)
     */
    suspend fun getAdvertisingId(): String? {
        ensureInitialized()
        return deviceInfo.getAdvertisingId()
    }
    
    /**
     * Start new session (automatically called on init)
     */
    fun startNewSession() {
        ensureInitialized()
        sessionId = UUID.randomUUID().toString()
        log("New session started: $sessionId")
    }
    
    internal fun getContext(): Context {
        ensureInitialized()
        return applicationContext
    }
    
    internal fun getConfig(): AdxConfig {
        ensureInitialized()
        return configuration
    }
    
    internal fun getNetworkClient(): NetworkClient {
        ensureInitialized()
        return networkClient
    }
    
    internal fun getDeviceInfo(): DeviceInfo {
        ensureInitialized()
        return deviceInfo
    }
    
    internal fun log(message: String, throwable: Throwable? = null) {
        if (::configuration.isInitialized && configuration.enableDebug) {
            if (throwable != null) {
                Log.d(TAG, message, throwable)
            } else {
                Log.d(TAG, message)
            }
        }
    }
    
    internal fun logError(message: String, throwable: Throwable? = null) {
        if (::configuration.isInitialized && configuration.enableDebug) {
            if (throwable != null) {
                Log.e(TAG, message, throwable)
            } else {
                Log.e(TAG, message)
            }
        }
    }
    
    private fun ensureInitialized() {
        check(isInitialized) {
            "AdxSDK is not initialized. Call AdxSDK.init() first."
        }
    }
}
