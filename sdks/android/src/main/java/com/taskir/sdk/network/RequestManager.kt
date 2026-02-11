package com.taskir.sdk.network

import android.util.Log
import com.taskir.sdk.data.models.ClientConfig
import com.taskir.sdk.data.models.ErrorResponse
import kotlinx.coroutines.delay
import okhttp3.Interceptor
import okhttp3.OkHttpClient
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.Response
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.util.UUID
import kotlin.math.pow

/**
 * HTTP Request Manager with retry logic
 * Handles all API communication with exponential backoff
 */
class RequestManager(private val config: ClientConfig) {

    private val okHttpClient: OkHttpClient by lazy {
        OkHttpClient.Builder()
            .addInterceptor(RequestInterceptor(config))
            .addInterceptor(logging = config.debug)
            .connectTimeout(config.timeout, java.util.concurrent.TimeUnit.MILLISECONDS)
            .readTimeout(config.timeout, java.util.concurrent.TimeUnit.MILLISECONDS)
            .writeTimeout(config.timeout, java.util.concurrent.TimeUnit.MILLISECONDS)
            .build()
    }

    val retrofit: Retrofit by lazy {
        Retrofit.Builder()
            .baseUrl(config.apiUrl)
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
    }

    /**
     * Execute API request with automatic retry logic
     */
    suspend inline fun <reified T> executeWithRetry(
        crossinline block: suspend () -> T
    ): T {
        var lastException: Exception? = null
        val maxAttempts = config.retryAttempts

        repeat(maxAttempts) { attempt ->
            try {
                return block()
            } catch (e: Exception) {
                lastException = e
                val isLastAttempt = attempt == maxAttempts - 1

                if (!isLastAttempt) {
                    val delayMs = calculateBackoffDelay(attempt)
                    log("Retry attempt ${attempt + 1}/$maxAttempts after ${delayMs}ms")
                    delay(delayMs)
                }
            }
        }

        throw lastException ?: Exception("Request failed after $maxAttempts attempts")
    }

    /**
     * Calculate exponential backoff delay
     */
    private fun calculateBackoffDelay(attemptNumber: Int): Long {
        val baseDelay = 100L
        return (baseDelay * 3.0.pow(attemptNumber.toDouble())).toLong()
    }

    private fun log(message: String) {
        if (config.debug) {
            Log.d("TaskirXSDK", message)
        }
    }

    private class RequestInterceptor(private val config: ClientConfig) : Interceptor {
        override fun intercept(chain: Interceptor.Chain): Response {
            val requestId = UUID.randomUUID().toString()
            val originalRequest = chain.request()

            val requestBuilder = originalRequest.newBuilder()
                .header("Content-Type", "application/json")
                .header("X-API-Key", config.apiKey)
                .header("X-Request-ID", requestId)
                .header("User-Agent", "TaskirX-Android-SDK/1.0")

            val request = requestBuilder.build()
            return chain.proceed(request)
        }
    }

    private fun OkHttpClient.Builder.addInterceptor(logging: Boolean): OkHttpClient.Builder {
        if (logging) {
            val interceptor = HttpLoggingInterceptor { message ->
                Log.d("TaskirXSDK", message)
            }
            interceptor.level = HttpLoggingInterceptor.Level.BODY
            addInterceptor(interceptor)
        }
        return this
    }

    companion object {
        private const val TAG = "RequestManager"
    }
}

/**
 * Custom exception for TaskirX SDK errors
 */
class TaskirXException(
    val code: String,
    val status: Int,
    message: String,
    val details: Map<String, Any>? = null,
    cause: Throwable? = null
) : Exception(message, cause)

/**
 * Exception mapper for HTTP responses
 */
object ExceptionMapper {
    fun mapError(status: Int, errorResponse: ErrorResponse): TaskirXException {
        val code = when (status) {
            400 -> "BAD_REQUEST"
            401 -> "UNAUTHORIZED"
            403 -> "FORBIDDEN"
            404 -> "NOT_FOUND"
            429 -> "RATE_LIMIT_EXCEEDED"
            500 -> "SERVER_ERROR"
            503 -> "SERVICE_UNAVAILABLE"
            else -> errorResponse.code
        }

        return TaskirXException(
            code = code,
            status = status,
            message = errorResponse.message,
            details = errorResponse.details
        )
    }
}

/**
 * HTTP Logging Interceptor
 */
class HttpLoggingInterceptor(private val logger: (String) -> Unit) : Interceptor {
    enum class Level {
        NONE, BASIC, HEADERS, BODY
    }

    var level = Level.NONE

    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()

        if (level == Level.NONE) {
            return chain.proceed(request)
        }

        logRequest(request)
        val startTime = System.nanoTime()
        val response = chain.proceed(request)
        val duration = System.nanoTime() - startTime

        logResponse(response, duration)
        return response
    }

    private fun logRequest(request: okhttp3.Request) {
        logger("--> ${request.method} ${request.url}")
        logger("Headers: ${request.headers}")
        request.body?.let {
            logger("Body: ${it.toString()}")
        }
    }

    private fun logResponse(response: Response, durationNanos: Long) {
        val durationMs = durationNanos / 1_000_000L
        logger("<-- ${response.code} (${durationMs}ms)")
        logger("Headers: ${response.headers}")
        response.body?.let {
            logger("Body: ${it.string()}")
        }
    }
}
