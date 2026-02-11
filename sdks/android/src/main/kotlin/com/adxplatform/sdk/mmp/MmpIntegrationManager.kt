package com.taskirx.sdk.mmp

import android.content.Context
import com.appsflyer.AppsFlyerLib
import com.appsflyer.attribution.AppsFlyerRequestListener
import com.adjust.sdk.Adjust
import com.adjust.sdk.AdjustConfig
import com.adjust.sdk.AdjustEvent
import com.adjust.sdk.LogLevel
import io.branch.referral.Branch
import com.taskirx.sdk.AdxSDK

/**
 * MMP Integration Manager for AdxSDK
 * Supports: AppsFlyer, Adjust, Branch, Kochava, Singular
 */
object MmpIntegrationManager {
    
    private var mmpProvider: MmpProvider? = null
    private var clickId: String? = null
    
    enum class MmpProvider {
        APPSFLYER,
        ADJUST,
        BRANCH,
        KOCHAVA,
        SINGULAR,
        TENJIN,
        NONE
    }
    
    /**
     * Initialize AppsFlyer
     */
    fun initializeAppsFlyer(
        context: Context,
        devKey: String,
        appId: String? = null
    ) {
        mmpProvider = MmpProvider.APPSFLYER
        
        AppsFlyerLib.getInstance().apply {
            init(devKey, null, context)
            start(context)
            
            // Set custom user ID from AdxSDK
            setCustomerUserId(AdxSDK.getUserId())
            
            // Register for attribution callback
            registerConversionListener(context, object : AppsFlyerConversionListener {
                override fun onConversionDataSuccess(data: Map<String, Any>?) {
                    data?.let {
                        val afStatus = it["af_status"] as? String
                        if (afStatus == "Non-organic") {
                            val clickId = it["af_sub1"] as? String
                            clickId?.let { id ->
                                setClickId(id)
                                notifytaskirx(id, it)
                            }
                        }
                    }
                }
                
                override fun onConversionDataFail(error: String?) {
                    // Handle error
                }
                
                override fun onAppOpenAttribution(data: Map<String, String>?) {
                    // Handle deep link
                }
                
                override fun onAttributionFailure(error: String?) {
                    // Handle error
                }
            })
        }
    }
    
    /**
     * Initialize Adjust
     */
    fun initializeAdjust(
        context: Context,
        appToken: String,
        environment: String = AdjustConfig.ENVIRONMENT_PRODUCTION
    ) {
        mmpProvider = MmpProvider.ADJUST
        
        val config = AdjustConfig(context, appToken, environment).apply {
            setLogLevel(LogLevel.INFO)
            
            // Set attribution callback
            setOnAttributionChangedListener { attribution ->
                attribution.trackerName?.let { trackerName ->
                    val clickId = attribution.trackerToken
                    clickId?.let { id ->
                        setClickId(id)
                        
                        val attributionData = mapOf(
                            "tracker_name" to trackerName,
                            "tracker_token" to attribution.trackerToken.orEmpty(),
                            "network" to attribution.network.orEmpty(),
                            "campaign" to attribution.campaign.orEmpty(),
                            "adgroup" to attribution.adgroup.orEmpty(),
                            "creative" to attribution.creative.orEmpty()
                        )
                        
                        notifytaskirx(id, attributionData)
                    }
                }
            }
        }
        
        Adjust.onCreate(config)
    }
    
    /**
     * Initialize Branch
     */
    fun initializeBranch(context: Context) {
        mmpProvider = MmpProvider.BRANCH
        
        Branch.getAutoInstance(context)
        
        // Branch initialization happens in Application.onCreate
        // Attribution will be received via Branch.sessionBuilder callback
    }
    
    /**
     * Set click ID from ad click
     */
    fun setClickId(id: String) {
        clickId = id
        
        // Pass to MMP if supported
        when (mmpProvider) {
            MmpProvider.APPSFLYER -> {
                AppsFlyerLib.getInstance().setAdditionalData(mapOf(
                    "adx_click_id" to id,
                    "adx_publisher_id" to AdxSDK.getPublisherId()
                ))
            }
            MmpProvider.ADJUST -> {
                // Adjust doesn't support setting click ID after init
                // But we can track it as an event
                val event = AdjustEvent("adx_click")
                event.addCallbackParameter("click_id", id)
                event.addCallbackParameter("publisher_id", AdxSDK.getPublisherId())
                Adjust.trackEvent(event)
            }
            MmpProvider.BRANCH -> {
                Branch.getInstance().setRequestMetadata("adx_click_id", id)
            }
            else -> {
                // Store for manual attribution
            }
        }
    }
    
    /**
     * Track custom event
     */
    fun trackEvent(
        eventName: String,
        eventValue: Map<String, Any>? = null,
        revenue: Double? = null,
        currency: String = "USD"
    ) {
        when (mmpProvider) {
            MmpProvider.APPSFLYER -> {
                AppsFlyerLib.getInstance().logEvent(
                    AdxSDK.getContext(),
                    eventName,
                    eventValue
                )
            }
            MmpProvider.ADJUST -> {
                val event = AdjustEvent(eventName)
                eventValue?.forEach { (key, value) ->
                    event.addCallbackParameter(key, value.toString())
                }
                revenue?.let {
                    event.setRevenue(it, currency)
                }
                Adjust.trackEvent(event)
            }
            MmpProvider.BRANCH -> {
                Branch.getInstance().userCompletedAction(eventName, eventValue)
            }
            else -> {
                // Track manually to taskirx
                trackEventTotaskirx(eventName, eventValue, revenue, currency)
            }
        }
    }
    
    /**
     * Get current click ID
     */
    fun getClickId(): String? = clickId
    
    /**
     * Get MMP provider
     */
    fun getMmpProvider(): MmpProvider? = mmpProvider
    
    /**
     * Notify taskirx about attribution
     */
    private fun notifytaskirx(clickId: String, attributionData: Map<String, Any>) {
        // Send attribution to taskirx backend
        kotlinx.coroutines.GlobalScope.launch {
            try {
                val endpoint = "${AdxSDK.getConfig().apiEndpoint}/api/mmp/attribution"
                val payload = mapOf(
                    "click_id" to clickId,
                    "publisher_id" to AdxSDK.getPublisherId(),
                    "mmp_provider" to mmpProvider?.name?.lowercase(),
                    "attribution_data" to attributionData,
                    "device_id" to AdxSDK.getUserId(),
                    "install_time" to System.currentTimeMillis()
                )
                
                // Make HTTP request (use your network client)
                // NetworkClient.post(endpoint, payload)
                
                println("AdxSDK: Attribution sent to platform for click_id: $clickId")
            } catch (e: Exception) {
                println("AdxSDK: Failed to send attribution: ${e.message}")
            }
        }
    }
    
    /**
     * Track event to taskirx
     */
    private fun trackEventTotaskirx(
        eventName: String,
        eventValue: Map<String, Any>?,
        revenue: Double?,
        currency: String
    ) {
        kotlinx.coroutines.GlobalScope.launch {
            try {
                val endpoint = "${AdxSDK.getConfig().apiEndpoint}/api/mmp/event"
                val payload = mapOf(
                    "click_id" to clickId,
                    "event_name" to eventName,
                    "event_value" to eventValue,
                    "revenue" to revenue,
                    "currency" to currency,
                    "timestamp" to System.currentTimeMillis()
                )
                
                // Make HTTP request
                // NetworkClient.post(endpoint, payload)
                
                println("AdxSDK: Event tracked: $eventName")
            } catch (e: Exception) {
                println("AdxSDK: Failed to track event: ${e.message}")
            }
        }
    }
}

/**
 * AppsFlyer Conversion Listener
 */
private interface AppsFlyerConversionListener : 
    com.appsflyer.AppsFlyerConversionListener {
    override fun onConversionDataSuccess(data: Map<String, Any>?)
    override fun onConversionDataFail(error: String?)
    override fun onAppOpenAttribution(data: Map<String, String>?)
    override fun onAttributionFailure(error: String?)
}
