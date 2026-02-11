package com.taskirx.sdk.internal

import com.taskirx.sdk.AdxSDK
import com.taskirx.sdk.config.AdxConfig
import com.taskirx.sdk.models.*
import com.squareup.moshi.Moshi
import com.squareup.moshi.kotlin.reflect.KotlinJsonAdapterFactory
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.logging.HttpLoggingInterceptor
import java.io.IOException
import java.util.concurrent.TimeUnit

/**
 * Network client for API requests
 */
internal class NetworkClient(private val config: AdxConfig) {
    
    private val client: OkHttpClient
    private val moshi: Moshi
    
    init {
        val logging = HttpLoggingInterceptor().apply {
            level = if (config.enableDebug) {
                HttpLoggingInterceptor.Level.BODY
            } else {
                HttpLoggingInterceptor.Level.NONE
            }
        }
        
        client = OkHttpClient.Builder()
            .connectTimeout(config.connectionTimeout, TimeUnit.SECONDS)
            .readTimeout(config.readTimeout, TimeUnit.SECONDS)
            .addInterceptor(logging)
            .addInterceptor { chain ->
                val request = chain.request().newBuilder()
                    .addHeader("Content-Type", "application/json")
                    .addHeader("X-ADX-SDK-Version", AdxSDK.getVersion())
                    .addHeader("X-ADX-Publisher-Id", AdxSDK.publisherId)
                    .build()
                chain.proceed(request)
            }
            .retryOnConnectionFailure(true)
            .build()
        
        moshi = Moshi.Builder()
            .add(KotlinJsonAdapterFactory())
            .build()
    }
    
    /**
     * Make RTB bid request
     */
    suspend fun requestAd(bidRequest: BidRequest): Result<BidResponse> = withContext(Dispatchers.IO) {
        try {
            val adapter = moshi.adapter(BidRequest::class.java)
            val json = adapter.toJson(bidRequest)
            
            val requestBody = json.toRequestBody("application/json".toMediaType())
            val request = Request.Builder()
                .url("${config.apiEndpoint}/api/rtb/bid-request")
                .post(requestBody)
                .build()
            
            AdxSDK.log("RTB Bid Request: $json")
            
            val response = client.newCall(request).execute()
            
            if (!response.isSuccessful) {
                val error = "HTTP ${response.code}: ${response.message}"
                AdxSDK.logError("Bid request failed: $error")
                return@withContext Result.failure(IOException(error))
            }
            
            val responseBody = response.body?.string()
            if (responseBody.isNullOrBlank()) {
                AdxSDK.logError("Empty response body")
                return@withContext Result.failure(IOException("Empty response"))
            }
            
            AdxSDK.log("RTB Bid Response: $responseBody")
            
            val bidResponse = moshi.adapter(BidResponse::class.java)
                .fromJson(responseBody)
            
            if (bidResponse == null) {
                AdxSDK.logError("Failed to parse bid response")
                return@withContext Result.failure(IOException("Failed to parse response"))
            }
            
            Result.success(bidResponse)
            
        } catch (e: Exception) {
            AdxSDK.logError("Network error: ${e.message}", e)
            Result.failure(e)
        }
    }
    
    /**
     * Track impression
     */
    suspend fun trackImpression(bidId: String): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            val request = Request.Builder()
                .url("${config.apiEndpoint}/api/rtb/impression/$bidId")
                .get()
                .build()
            
            val response = client.newCall(request).execute()
            
            if (response.isSuccessful) {
                AdxSDK.log("Impression tracked: $bidId")
                Result.success(Unit)
            } else {
                Result.failure(IOException("HTTP ${response.code}"))
            }
        } catch (e: Exception) {
            AdxSDK.logError("Failed to track impression: ${e.message}", e)
            Result.failure(e)
        }
    }
    
    /**
     * Track click
     */
    suspend fun trackClick(bidId: String): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            val request = Request.Builder()
                .url("${config.apiEndpoint}/api/rtb/click/$bidId")
                .get()
                .build()
            
            val response = client.newCall(request).execute()
            
            if (response.isSuccessful) {
                AdxSDK.log("Click tracked: $bidId")
                Result.success(Unit)
            } else {
                Result.failure(IOException("HTTP ${response.code}"))
            }
        } catch (e: Exception) {
            AdxSDK.logError("Failed to track click: ${e.message}", e)
            Result.failure(e)
        }
    }
    
    /**
     * Track conversion
     */
    suspend fun trackConversion(
        bidId: String,
        eventType: String,
        value: Double? = null
    ): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            val url = "${config.apiEndpoint}/api/rtb/conversion/$bidId" +
                    "?eventType=$eventType" +
                    (value?.let { "&value=$it" } ?: "")
            
            val request = Request.Builder()
                .url(url)
                .get()
                .build()
            
            val response = client.newCall(request).execute()
            
            if (response.isSuccessful) {
                AdxSDK.log("Conversion tracked: $bidId ($eventType)")
                Result.success(Unit)
            } else {
                Result.failure(IOException("HTTP ${response.code}"))
            }
        } catch (e: Exception) {
            AdxSDK.logError("Failed to track conversion: ${e.message}", e)
            Result.failure(e)
        }
    }
}
